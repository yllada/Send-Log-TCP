package stress_test

import (
	"net"
	"testing"
	"time"

	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/stress"
	"Sendlog-Syslog/internal/syslog"
)

// TestEngineOverLiveClientConn mimics the GUI: dial a persistent client,
// hand its connection to the stress engine, and verify several framed
// messages arrive on ONE connection. This is the "connected stress" path.
func TestEngineOverLiveClientConn(t *testing.T) {
	server, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = server.Close() }()
	host, port, err := net.SplitHostPort(server.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	var frames int
	go func() {
		conn, err := server.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		for _, f := range readFrames(t, conn, 5) {
			_ = f
			frames++
		}
	}()

	client, err := syslog.DialClient(syslog.Config{
		Address: host, Port: port, Protocol: "tcp",
		Method: framing.OctetCounting, Facility: 13, Severity: 6,
	})
	if err != nil {
		t.Fatalf("DialClient: %v", err)
	}
	defer func() { _ = client.Close() }()

	conn, err := client.Connection()
	if err != nil {
		t.Fatalf("Connection: %v", err)
	}

	eng := stress.NewEngine()
	cfg := stress.ContinuousSendConfig{
		Address:        host,
		Port:           port,
		Protocol:       "tcp",
		Method:         framing.OctetCounting,
		Facility:       13,
		Severity:       6,
		Message:        "connected stress",
		MessagesPerSec: 100,
		MaxMessages:    5,
		Conn:           conn,
	}
	if err := eng.Start(cfg, nil); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := engineStatsSent(t, eng, 5, 3*time.Second); got < 5 {
		t.Fatalf("engine reports sent=%d, want >=5", got)
	}
	eng.Stop()

	// The client connection must still be usable after stress (no close).
	resp := client.SendMessages(syslog.Config{Facility: 13, Severity: 6}, []string{"after-stress"}, nil)
	if len(resp.Errors) != 0 || len(resp.SentMessages) != 1 {
		t.Errorf("post-stress send on same conn: sent=%d err=%v; want usable",
			len(resp.SentMessages), resp.Errors)
	}
}
