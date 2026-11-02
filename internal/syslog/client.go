package syslog

import (
	"fmt"
	"net"
	"sync"
	"time"

	"Sendlog-Syslog/internal/framing"
)

// Client holds one transport connection reused across sends. Facility,
// severity, and message format are chosen per SendMessages call, so the GUI
// can change them after connecting. Senders that want a fresh connection per
// operation keep using Dial.
type Client struct {
	mu     sync.Mutex
	conn   net.Conn
	framer *framing.Framer // non-nil only for TCP
	closed bool
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

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		response.Errors = append(response.Errors, "client is closed; reconnect first")
		return response
	}

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

		_ = c.conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
		if err := WriteAll(c.conn, payload); err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("send: %v", err))
			break
		}
		response.SentMessages = append(response.SentMessages, syslogMsg)
		c.reportProgress(i, len(messages), onProgress)
	}

	return response
}

// Connection exposes the underlying socket so callers (the stress engine) can
// write to it directly. It is shared: only one writer at a time.
func (c *Client) Connection() (net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, fmt.Errorf("client is closed; reconnect first")
	}
	return c.conn, nil
}

// Close releases the transport. Further SendMessages calls record an error.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

func (c *Client) reportProgress(index, total int, onProgress SendFunc) {
	if onProgress != nil {
		onProgress(index+1, total, float64(index+1)/float64(total)*100)
	}
}
