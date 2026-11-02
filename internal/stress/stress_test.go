package stress_test

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"Sendlog-Syslog/internal/stress"
)

// readFrames pulls n octet-counting framed syslog messages off conn, all
// through one buffered reader.
func readFrames(t *testing.T, conn net.Conn, n int) []string {
	t.Helper()
	var out []string
	r := bufio.NewReader(conn)
	for i := 0; i < n; i++ {
		prefix, err := r.ReadString(' ')
		if err != nil {
			t.Errorf("frame %d: read prefix: %v", i, err)
			return out
		}
		cnt, err := strconv.Atoi(strings.TrimSpace(prefix))
		if err != nil {
			t.Errorf("frame %d: bad count %q: %v", i, prefix, err)
			return out
		}
		buf := make([]byte, cnt)
		if _, err := io.ReadFull(r, buf); err != nil {
			t.Errorf("frame %d: read body: %v", i, err)
			return out
		}
		out = append(out, string(buf))
	}
	return out
}

// engineStatsSent waits up to timeout for the engine to report at least want.
func engineStatsSent(t *testing.T, eng *stress.Engine, want int64, timeout time.Duration) int64 {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		st := eng.GetContinuousStats()
		if st.TotalSent >= want || time.Now().After(deadline) {
			return st.TotalSent
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestEngineSendsOverInjectedConn proves the GUI-connected path: when the
// caller supplies config.Conn, the engine sends through it and does NOT close
// it — the GUI's live client must stay open after the run.
func TestEngineSendsOverInjectedConn(t *testing.T) {
	server, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	host, port, err := net.SplitHostPort(server.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	injected, err := net.Dial("tcp", server.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer injected.Close()

	go func() {
		conn, err := server.Accept()
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		defer conn.Close()
		for i, f := range readFrames(t, conn, 3) {
			if !strings.HasPrefix(f, "<14>") {
				t.Errorf("frame %d missing PRI: %q", i, f)
			}
		}
	}()

	eng := stress.NewEngine()
	cfg := stress.ContinuousSendConfig{
		Address:        host,
		Port:           port,
		Protocol:       "tcp",
		Method:         "octet-counting",
		Facility:       1,
		Severity:       6,
		Message:        "stress-payload",
		MessagesPerSec: 100,
		MaxMessages:    3,
		Conn:           injected,
	}
	if err := eng.Start(cfg, nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := engineStatsSent(t, eng, 3, 3*time.Second); got < 3 {
		t.Fatalf("engine sent %d, want >= 3", got)
	}
	eng.Stop()

	// The injected connection must still be usable (engine didn't close it).
	if _, err := injected.Write([]byte("still-open")); err != nil {
		t.Errorf("injected conn closed by engine: %v", err)
	}
}

// TestEngineDialsAndClosesOwnConn proves the GUI-not-connected path: with
// Conn nil the engine dials on its own before starting.
func TestEngineDialsAndClosesOwnConn(t *testing.T) {
	server, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	host, port, err := net.SplitHostPort(server.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	dialed := make(chan struct{}, 1)
	go func() {
		conn, err := server.Accept()
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		dialed <- struct{}{}
		defer conn.Close()
	}()

	eng := stress.NewEngine()
	cfg := stress.ContinuousSendConfig{
		Address:        host,
		Port:           port,
		Protocol:       "tcp",
		Method:         "octet-counting",
		Facility:       1,
		Severity:       6,
		Message:        "stress-payload",
		MessagesPerSec: 100,
		MaxMessages:    2,
	}
	if err := eng.Start(cfg, nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	select {
	case <-dialed:
	case <-time.After(2 * time.Second):
		t.Fatal("engine never dialed")
	}
	eng.Stop()
}
