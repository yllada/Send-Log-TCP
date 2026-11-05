package syslog

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// oneShotServer accepts a connection, reads exactly one octet-counted frame,
// then closes the connection — mimicking a syslog server that drops the TCP
// link after every message. It repeats for want frames, counting each into
// count. The client must re-dial after every close.
func oneShotServer(t *testing.T, want int, count *int32) (string, string) {
	t.Helper()
	l, host, port := startTestListener(t, "tcp")
	go func() {
		for atomic.LoadInt32(count) < int32(want) {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			r := bufio.NewReader(conn)
			prefix, err := r.ReadString(' ')
			if err != nil {
				_ = conn.Close()
				return
			}
			n, err := strconv.Atoi(strings.TrimSpace(prefix))
			if err != nil {
				_ = conn.Close()
				return
			}
			buf := make([]byte, n)
			if _, err := io.ReadFull(r, buf); err != nil {
				_ = conn.Close()
				return
			}
			atomic.AddInt32(count, 1)
			_ = conn.Close()
		}
	}()
	return host, port
}

func waitForFrames(t *testing.T, count *int32, want int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for atomic.LoadInt32(count) < int32(want) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if got := atomic.LoadInt32(count); got != int32(want) {
		t.Errorf("one-shot server saw %d frames, want %d", got, want)
	}
}

// TestClientAutoReconnectsPerSend proves the reported fix: when the server
// drops the TCP connection after each message, repeated single-message sends
// keep succeeding without a manual disconnect/reconnect.
func TestClientAutoReconnectsPerSend(t *testing.T) {
	var count int32
	host, port := oneShotServer(t, 3, &count)

	client, err := DialClient(Config{
		Address: host, Port: port, Protocol: "tcp",
		Method: "octet-counting", Facility: 13, Severity: 6,
	})
	if err != nil {
		t.Fatalf("DialClient: %v", err)
	}
	defer func() { _ = client.Close() }()

	for i, msg := range []string{"one", "two", "three"} {
		resp := client.SendMessages(Config{Facility: 13, Severity: 6}, []string{msg}, nil)
		if len(resp.Errors) != 0 || len(resp.SentMessages) != 1 {
			t.Fatalf("send %d (%s): sent=%d err=%v; want 1 sent, no errors",
				i+1, msg, len(resp.SentMessages), resp.Errors)
		}
	}

	waitForFrames(t, &count, 3)
}

// TestClientAutoReconnectsBatch proves a whole multi-message batch survives a
// server that drops the connection after every single frame.
func TestClientAutoReconnectsBatch(t *testing.T) {
	var count int32
	host, port := oneShotServer(t, 4, &count)

	client, err := DialClient(Config{
		Address: host, Port: port, Protocol: "tcp",
		Method: "octet-counting", Facility: 13, Severity: 6,
	})
	if err != nil {
		t.Fatalf("DialClient: %v", err)
	}
	defer func() { _ = client.Close() }()

	resp := client.SendMessages(Config{Facility: 13, Severity: 6},
		[]string{"a", "b", "c", "d"}, nil)
	if len(resp.Errors) != 0 || len(resp.SentMessages) != 4 {
		t.Fatalf("batch: sent=%d err=%v; want 4 sent, no errors",
			len(resp.SentMessages), resp.Errors)
	}

	waitForFrames(t, &count, 4)
}
