package syslog

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func selfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyBytes, _ := x509.MarshalECPrivateKey(key)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

func TestClientTLSPersistent(t *testing.T) {
	cert := selfSignedCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	l, err := tls.Listen("tcp", "127.0.0.1:0", tlsCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	host, port, _ := net.SplitHostPort(l.Addr().String())

	var count int32
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
			atomic.AddInt32(&count, 1)
		}
	}()

	client, err := DialClient(Config{
		Address: host, Port: port, Protocol: "tcp",
		Method: "octet-counting", Facility: 13, Severity: 6,
		UseTLS: true, TLSVerify: false,
	})
	if err != nil {
		t.Fatalf("DialClient TLS: %v", err)
	}
	defer func() { _ = client.Close() }()

	for i, msg := range []string{"tls-one", "tls-two", "tls-three"} {
		resp := client.SendMessages(Config{Facility: 13, Severity: 6, UseTLS: true, TLSVerify: false}, []string{msg}, nil)
		if len(resp.Errors) != 0 || len(resp.SentMessages) != 1 {
			t.Errorf("send %d (%s): sent=%d err=%v", i+1, msg, len(resp.SentMessages), resp.Errors)
		}
	}

	deadline := time.Now().Add(2 * time.Second)
	for atomic.LoadInt32(&count) < 3 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if got := atomic.LoadInt32(&count); got != 3 {
		t.Errorf("TLS server received %d frames, want 3", got)
	}
}
