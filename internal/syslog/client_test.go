package syslog

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"Sendlog-Syslog/internal/framing"
)

func startTestListener(t *testing.T, protocol string) (net.Listener, string, string) {
	t.Helper()
	l, err := net.Listen(protocol, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen %s: %v", protocol, err)
	}
	t.Cleanup(func() { _ = l.Close() })
	host, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		t.Fatalf("split addr: %v", err)
	}
	return l, host, port
}

func TestClientSend_OctetCountingTCP(t *testing.T) {
	l, host, port := startTestListener(t, "tcp")
	dials := int32(0)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		atomic.AddInt32(&dials, 1)
		defer func() { _ = conn.Close() }()
		reader := bufio.NewReader(conn)
		for i := 0; i < 2; i++ {
			prefix, err := reader.ReadString(' ')
			if err != nil {
				return
			}
			n, err := strconv.Atoi(strings.TrimSpace(prefix))
			if err != nil {
				return
			}
			buf := make([]byte, n)
			if _, err := io.ReadFull(reader, buf); err != nil {
				return
			}
			if !strings.HasPrefix(string(buf), "<14>") {
				t.Errorf("message %d lacks <14> PRI: %q", i+1, buf)
			}
		}
	}()

	client, err := DialClient(Config{
		Address: host, Port: port, Protocol: "tcp",
		Method: framing.OctetCounting, Facility: 1, Severity: 6,
	})
	if err != nil {
		t.Fatalf("DialClient: %v", err)
	}

	resp := client.SendMessages(Config{Facility: 1, Severity: 6}, []string{"first", "second"}, nil)
	if len(resp.SentMessages) != 2 || len(resp.Errors) != 0 {
		t.Fatalf("SendMessages = %d sent, %d errors; want 2 sent, 0 errors", len(resp.SentMessages), len(resp.Errors))
	}
	_ = client.Close()
	if got := atomic.LoadInt32(&dials); got != 1 {
		t.Fatalf("server saw %d dials, want 1", got)
	}
}

func TestClientCloseRejectsFurtherSend(t *testing.T) {
	l, host, port := startTestListener(t, "tcp")
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		buf := make([]byte, 64)
		_, _ = conn.Read(buf)
	}()

	client, err := DialClient(Config{
		Address: host, Port: port, Protocol: "tcp",
		Method: framing.OctetCounting, Facility: 1, Severity: 6,
	})
	if err != nil {
		t.Fatalf("DialClient: %v", err)
	}
	_ = client.Close()

	resp := client.SendMessages(Config{Facility: 1, Severity: 6}, []string{"after close"}, nil)
	if len(resp.SentMessages) != 0 || len(resp.Errors) != 1 {
		t.Fatalf("send after close = %d sent, %d errors; want 0 sent, 1 error", len(resp.SentMessages), len(resp.Errors))
	}
}

func TestClientSend_UDPPerMessage(t *testing.T) {
	udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("resolve udp addr: %v", err)
	}
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	t.Cleanup(func() { _ = udpConn.Close() })
	port := udpConn.LocalAddr().(*net.UDPAddr).Port
	received := make(chan string, 1)
	go func() {
		buf := make([]byte, 1500)
		n, _, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		received <- string(buf[:n])
	}()

	client, err := DialClient(Config{Address: "127.0.0.1", Port: strconv.Itoa(port), Protocol: "udp", Facility: 1, Severity: 6})
	if err != nil {
		t.Fatalf("DialClient: %v", err)
	}
	defer func() { _ = client.Close() }()

	resp := client.SendMessages(Config{Facility: 1, Severity: 6}, []string{"hello"}, nil)
	if len(resp.SentMessages) != 1 {
		t.Fatalf("SendMessages = %d sent; want 1", len(resp.SentMessages))
	}
	select {
	case msg := <-received:
		if !strings.HasPrefix(msg, "<14>") {
			t.Errorf("UDP datagram lacks <14> PRI: %q", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no UDP datagram received")
	}
}

func TestDialClient_InvalidConfigRejected(t *testing.T) {
	if _, err := DialClient(Config{Port: "514"}); err == nil {
		t.Fatal("DialClient accepted empty address")
	}
	if _, err := DialClient(Config{
		Address: "127.0.0.1", Port: "514", Protocol: "udp",
		UseTLS: true,
	}); err == nil {
		t.Fatal("DialClient accepted TLS over UDP")
	}
}
