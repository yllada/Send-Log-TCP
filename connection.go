// Package main implements the Send-Log-TCP desktop application: a Wails-based
// GUI for composing and sending syslog messages over TCP, UDP and TLS.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"Sendlog-Syslog/internal/syslog"
)

// ================================================================================
// CONNECTION SERVICE - GUI-facing connection manager
// The dialing logic itself lives in internal/syslog; this service only keeps
// the Wails runtime context and the active connection for the GUI.
// ================================================================================

// ConnectionService manages network connections to syslog servers
type ConnectionService struct {
	ctx         context.Context
	conn        net.Conn
	connMu      sync.Mutex
	isConnected bool
}

// NewConnectionService creates a new ConnectionService instance
func NewConnectionService() *ConnectionService {
	return &ConnectionService{}
}

// SetContext sets the Wails runtime context
func (c *ConnectionService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// IsConnected returns the current connection state
func (c *ConnectionService) IsConnected() bool {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.isConnected
}

// GetConnection returns the current connection (for use by other services)
func (c *ConnectionService) GetConnection() net.Conn {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn
}

// CheckConnection verifies if a connection can be established
// Supports TCP, UDP, and TCP+TLS (RFC 5425 for secure syslog)
func (c *ConnectionService) CheckConnection(address string, port string, protocol string, useTLS bool, tlsVerify bool, caCertPath string, clientCertPath string, clientKeyPath string) (bool, error) {
	fullAddress := net.JoinHostPort(address, port)

	conn, err := syslog.Dial(syslog.Config{
		Address:        address,
		Port:           port,
		Protocol:       protocol,
		UseTLS:         useTLS,
		TLSVerify:      tlsVerify,
		CACertPath:     caCertPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
	})
	if err != nil {
		runtime.LogError(c.ctx, fmt.Sprintf("Error connecting to %s: %v", fullAddress, err))
		return false, err
	}

	// Store the active connection
	c.connMu.Lock()
	c.conn = conn
	c.isConnected = true
	c.connMu.Unlock()

	protocolInfo := protocol
	if useTLS {
		protocolInfo = fmt.Sprintf("%s+TLS", protocolInfo)
	}
	runtime.LogInfo(c.ctx, fmt.Sprintf("Connected to %s via %s", fullAddress, protocolInfo))
	return true, nil
}

// Disconnect closes the active connection if it exists
func (c *ConnectionService) Disconnect() {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.isConnected {
		err := c.conn.Close()
		if err != nil {
			log.Printf("Error disconnecting: %v", err)
		} else {
			log.Println("Disconnected successfully.")
		}
		c.isConnected = false
	}
}

// SelectCACertificate opens a dialog to select a CA certificate file
func (c *ConnectionService) SelectCACertificate() (string, error) {
	filePath, err := runtime.OpenFileDialog(c.ctx, runtime.OpenDialogOptions{
		Title: "Select CA Certificate",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Certificate Files (*.pem, *.crt, *.cer)",
				Pattern:     "*.pem;*.crt;*.cer;*.ca-bundle",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// SelectClientCertificate opens a dialog to select a client certificate (mTLS)
func (c *ConnectionService) SelectClientCertificate() (string, error) {
	filePath, err := runtime.OpenFileDialog(c.ctx, runtime.OpenDialogOptions{
		Title: "Select Client Certificate",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Certificate Files (*.pem, *.crt, *.cer)",
				Pattern:     "*.pem;*.crt;*.cer",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// SelectClientKey opens a dialog to select a client private key
func (c *ConnectionService) SelectClientKey() (string, error) {
	filePath, err := runtime.OpenFileDialog(c.ctx, runtime.OpenDialogOptions{
		Title: "Select Client Key",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Key Files (*.pem, *.key)",
				Pattern:     "*.pem;*.key",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}
