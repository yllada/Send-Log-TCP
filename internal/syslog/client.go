package syslog

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"Sendlog-Syslog/internal/framing"
)

// Client holds one transport connection reused across sends. Facility,
// severity, and message format are chosen per SendMessages call, so the GUI
// can change them after connecting.
//
// For TCP (and TLS-over-TCP) the client is resilient: if the server drops the
// connection after a write, the client transparently re-dials once and retries
// the message, so callers do not have to manually disconnect and reconnect.
// UDP is connectionless and never re-dials.
type Client struct {
	writeMu sync.Mutex // serializes every socket write
	connMu  sync.Mutex // guards conn, closed, and dialCfg

	conn   net.Conn
	closed bool
	// framer is non-nil only for TCP.
	framer *framing.Framer
	// dialCfg is the transport used to re-dial after a dropped TCP connection;
	// nil for UDP.
	dialCfg *DialConfig
}

// DialClient validates the config and opens the transport. The caller owns the
// returned client and must Close it.
func DialClient(config Config) (*Client, error) {
	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	conn, err := Dial(config)
	if err != nil {
		return nil, fmt.Errorf("connect to %s:%s: %w", config.Address, config.Port, err)
	}

	c := &Client{conn: conn}
	if config.Protocol == "tcp" {
		c.framer = framing.NewFramer(framing.Config{
			Method:           config.Method,
			ValidateUTF8:     true,
			MaxMessageLength: 0,
		})
		c.dialCfg = &DialConfig{
			Address:        config.Address,
			Port:           config.Port,
			Protocol:       config.Protocol,
			UseTLS:         config.UseTLS,
			TLSVerify:      config.TLSVerify,
			CACertPath:     config.CACertPath,
			ClientCertPath: config.ClientCertPath,
			ClientKeyPath:  config.ClientKeyPath,
		}
	}
	return c, nil
}

// SendMessages writes every message over the open transport using the
// connection's framing for the transport. Per-message fields (facility,
// severity, format) come from config. A broken or closed transport records an
// error and stops.
func (c *Client) SendMessages(config Config, messages []string, onProgress SendFunc) Response {
	var response Response
	response.SentMessages = []string{}
	response.Errors = []string{}

	c.connMu.Lock()
	if c.closed {
		c.connMu.Unlock()
		response.Errors = append(response.Errors, "client is closed; reconnect first")
		return response
	}
	c.connMu.Unlock()

	for i, message := range messages {
		syslogMsg, err := BuildMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("build: %v", err))
			c.reportProgress(i, len(messages), onProgress)
			continue
		}

		payload := []byte(syslogMsg)
		if c.framer != nil {
			if payload, err = c.framer.Frame(syslogMsg); err != nil {
				response.Errors = append(response.Errors, fmt.Sprintf("frame: %v", err))
				c.reportProgress(i, len(messages), onProgress)
				continue
			}
		}

		if _, err := c.write(payload); err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("send: %v", err))
			break
		}
		response.SentMessages = append(response.SentMessages, syslogMsg)
		c.reportProgress(i, len(messages), onProgress)
	}

	return response
}

// Connection returns a stable net.Conn handle bound to this client. Its Write
// goes through the client's auto-reconnecting write, so a dropped TCP
// connection is recovered transparently even for long-running writers such as
// the stress engine. Read and the other net.Conn methods delegate to the
// current transport.
func (c *Client) Connection() (net.Conn, error) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.closed {
		return nil, errors.New("client is closed; reconnect first")
	}
	return &clientConn{c: c}, nil
}

// Close releases the transport. Further SendMessages calls record an error.
func (c *Client) Close() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

// write writes b over the current transport, serializing with other writers.
// For a dialable TCP connection it first probes the socket: a peer that closed
// it (some syslog servers drop the link after each message) is re-dialed
// before the write, so no message is lost into a dead socket. A write that
// still fails because the connection dropped re-dials once and retries b; any
// other error is returned.
func (c *Client) write(b []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	conn, closed, dialCfg := c.snapshot()
	if closed {
		return 0, errors.New("client is closed; reconnect first")
	}
	if conn == nil {
		return 0, errors.New("client has no connection")
	}
	if dialCfg != nil && connDropped(conn) {
		fresh, err := dialCfg.Dial()
		if err != nil {
			return 0, err
		}
		c.swapConn(conn, fresh)
		conn = fresh
	}

	err := writeWithDeadline(conn, b)
	if err == nil {
		return len(b), nil
	}
	// The write failed because the peer dropped the connection: re-dial once
	// and retry.
	if dialCfg != nil && isConnectionDropped(err) {
		if fresh, derr := dialCfg.Dial(); derr == nil {
			c.swapConn(conn, fresh)
			if werr := writeWithDeadline(fresh, b); werr == nil {
				return len(b), nil
			}
		}
	}
	return 0, err
}

// connDropped reports whether the peer has closed conn. It performs a brief
// read: a dropped connection yields EOF or a reset as soon as the peer's FIN
// is queued, while an alive connection yields a deadline timeout (syslog
// servers send nothing on the link). The read deadline is cleared afterwards.
func connDropped(conn net.Conn) bool {
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Millisecond)); err != nil {
		return false
	}
	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	_ = conn.SetReadDeadline(time.Time{})
	if err == nil {
		return false
	}
	return errors.Is(err, io.EOF) || isConnectionDropped(err)
}

// snapshot reads the current transport, closed flag, and re-dial config
// atomically.
func (c *Client) snapshot() (net.Conn, bool, *DialConfig) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn, c.closed, c.dialCfg
}

// swapConn atomically replaces old with fresh and closes old best-effort.
func (c *Client) swapConn(old, fresh net.Conn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn == old {
		c.conn = fresh
		_ = old.Close()
	}
}

// writeWithDeadline writes all of b to conn within WriteTimeout, delegating
// to the exported WriteAll so there is a single partial-write loop.
func writeWithDeadline(conn net.Conn, b []byte) error {
	_ = conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
	return WriteAll(conn, b)
}

// isConnectionDropped reports whether err indicates the peer closed or reset
// the connection mid-write, which is the case worth transparently re-dialing.
func isConnectionDropped(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, marker := range []string{
		"eof",
		"connection reset",
		"broken pipe",
		"reset by peer",
		"peer closed",
		"remote end hung up",
		"use of closed network connection",
		"connection closed by server",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// clientConn is a net.Conn handle bound to a *Client. Writes go through the
// client's auto-reconnecting write; every other method delegates to the
// client's current transport at call time.
type clientConn struct {
	c *Client
}

func (cc *clientConn) Write(b []byte) (int, error) { return cc.c.write(b) }

func (cc *clientConn) Read(b []byte) (int, error) {
	conn, _, _ := cc.c.snapshot()
	if conn == nil {
		return 0, errors.New("client has no connection")
	}
	return conn.Read(b)
}

func (cc *clientConn) Close() error { return cc.c.Close() }

func (cc *clientConn) LocalAddr() net.Addr {
	conn, _, _ := cc.c.snapshot()
	if conn == nil {
		return nil
	}
	return conn.LocalAddr()
}

func (cc *clientConn) RemoteAddr() net.Addr {
	conn, _, _ := cc.c.snapshot()
	if conn == nil {
		return nil
	}
	return conn.RemoteAddr()
}

func (cc *clientConn) SetDeadline(t time.Time) error {
	conn, _, _ := cc.c.snapshot()
	if conn == nil {
		return errors.New("client has no connection")
	}
	return conn.SetDeadline(t)
}

func (cc *clientConn) SetReadDeadline(t time.Time) error {
	conn, _, _ := cc.c.snapshot()
	if conn == nil {
		return errors.New("client has no connection")
	}
	return conn.SetReadDeadline(t)
}

func (cc *clientConn) SetWriteDeadline(t time.Time) error {
	conn, _, _ := cc.c.snapshot()
	if conn == nil {
		return errors.New("client has no connection")
	}
	return conn.SetWriteDeadline(t)
}

func (c *Client) reportProgress(index, total int, onProgress SendFunc) {
	if onProgress != nil {
		onProgress(index+1, total, float64(index+1)/float64(total)*100)
	}
}
