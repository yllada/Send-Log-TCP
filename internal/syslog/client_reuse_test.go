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

// persistentServer accepts a single TCP connection and keeps reading
// octet-counted frames from it, counting each. Unlike a real syslog server it
// never drops the link, so it proves the client reuses one connection.
func persistentServer(t *testing.T) (*int32, string, string) {
	t.Helper()
	l, host, port := startTestListener(t, "tcp")
	count := new(int32)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		r := bufio.NewReader(conn)
		for {
			prefix, err := r.ReadString(' ')
			if err != nil {
				return
			}
			n, err := strconv.Atoi(strings.TrimSpace(prefix))
			if err != nil {
				return
			}
			buf := make([]byte, n)
			if _, err := io.ReadFull(r, buf); err != nil {
				return
			}
			atomic.AddInt32(count, 1)
		}
	}()
	return count, host, port
}

// TestClientReusedSend proves the reported flow: connect once, send several
// single-message sends on the same Client without reconnecting.
func TestClientReusedSend(t *testing.T) {
	count, host, port := persistentServer(t)

	client, err := DialClient(Config{
		Address: host, Port: port, Protocol: "tcp",
		Method: "octet-counting", Facility: 13, Severity: 6,
	})
	if err != nil {
		t.Fatalf("DialClient: %v", err)
	}
	defer func() { _ = client.Close() }()

	for _, msg := range []string{"one", "two", "three"} {
		resp := client.SendMessages(Config{Facility: 13, Severity: 6}, []string{msg}, nil)
		if len(resp.Errors) != 0 || len(resp.SentMessages) != 1 {
			t.Fatalf("send %q: sent=%d err=%v; want 1 sent, no errors",
				msg, len(resp.SentMessages), resp.Errors)
		}
	}

	deadline := time.Now().Add(2 * time.Second)
	for atomic.LoadInt32(count) < 3 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if got := atomic.LoadInt32(count); got != 3 {
		t.Errorf("server received %d frames after 3 sends on one client; want 3", got)
	}
}
