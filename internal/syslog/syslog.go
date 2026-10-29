// Package syslog implements the SendLog core: configuration validation,
// RFC 5424 / RFC 3164 / raw-pri message building, and sending over TCP,
// UDP or TLS. It is independent of the Wails GUI and is shared by the
// desktop application and the CLI.
package syslog

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"Sendlog-Syslog/internal/framing"
)

// Connection timeouts according to best practices
const (
	ConnectionTimeout = 10 * time.Second
	WriteTimeout      = 30 * time.Second
)

// MessageFormat specifies how the syslog message is assembled before framing
type MessageFormat string

const (
	// MessageFormatRFC5424 builds a full RFC 5424 header (version, timestamp,
	// hostname, app-name, nil procid/msgid/structured-data) around the message
	MessageFormatRFC5424 MessageFormat = "rfc5424"

	// MessageFormatRFC3164 builds a legacy BSD syslog header (RFC 3164) around
	// the message
	MessageFormatRFC3164 MessageFormat = "rfc3164"

	// MessageFormatRawPRI emits only "<PRI>" followed by the message bytes
	// verbatim: no version digit, no space, no timestamp, no hostname, no
	// app-name, no BOM and no trailing newline.
	//
	// This exists to replay real vendor logs. A captured line (for example an
	// ESET ESMC event) already carries its own complete RFC 5424 header, so
	// injecting another one would corrupt the event. The operator supplies
	// everything after the PRI.
	MessageFormatRawPRI MessageFormat = "raw-pri"
)

// IsValidMessageFormat reports whether the message format is recognized
func IsValidMessageFormat(format MessageFormat) bool {
	switch format {
	case MessageFormatRFC5424, MessageFormatRFC3164, MessageFormatRawPRI:
		return true
	default:
		return false
	}
}

// ResolveMessageFormat determines the effective message format.
//
// A recognized, non-empty format always wins. When the format is absent (or
// not recognized) it falls back to the legacy UseRFC5424 boolean, which keeps
// templates stored before MessageFormat existed working without any storage
// migration.
func ResolveMessageFormat(format MessageFormat, useRFC5424 bool) MessageFormat {
	if IsValidMessageFormat(format) {
		return format
	}
	if useRFC5424 {
		return MessageFormatRFC5424
	}
	return MessageFormatRFC3164
}

// Config holds the configuration for sending syslog messages
type Config struct {
	Address       string         `json:"Address"`
	Port          string         `json:"Port"`
	Protocol      string         `json:"Protocol"`
	Messages      []string       `json:"Messages"`
	Method        framing.Method `json:"FramingMethod"`
	Facility      uint8          `json:"Facility"`
	Severity      uint8          `json:"Severity"`
	Hostname      string         `json:"Hostname"`
	Appname       string         `json:"Appname"`
	MessageFormat MessageFormat  `json:"MessageFormat"`
	// UseRFC5424 is the legacy format switch. It is kept because stored
	// templates and profiles persist it; MessageFormat takes precedence when set.
	UseRFC5424     bool   `json:"UseRFC5424"`
	UseTLS         bool   `json:"UseTLS"`
	TLSVerify      bool   `json:"TLSVerify"`
	CACertPath     string `json:"CACertPath"`
	ClientCertPath string `json:"ClientCertPath"`
	ClientKeyPath  string `json:"ClientKeyPath"`
}

// Response contains the result of send operations
type Response struct {
	SentMessages []string `json:"sentMessages"`
	Errors       []string `json:"errors"`
}

// SendFunc is an optional callback invoked after each message; nil means
// no progress reporting. The GUI wires its runtime events here; the CLI
// wires stderr printing.
type SendFunc func(current, total int, percent float64)

// SendSyslogMessages sends syslog messages and returns the result.
func SendSyslogMessages(config Config, onProgress SendFunc) Response {
	response := Response{
		SentMessages: []string{},
		Errors:       []string{},
	}

	// Validate configuration
	if err := ValidateConfig(&config); err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Invalid configuration: %v", err))
		return response
	}

	// Establish connection
	conn, err := Dial(config)
	if err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Error connecting to %s:%s: %v", config.Address, config.Port, err))
		return response
	}
	defer func() { _ = conn.Close() }()

	// Send messages based on protocol
	if config.Protocol == "tcp" {
		return sendTCPMessages(conn, config, onProgress)
	}
	return sendUDPMessages(conn, config, onProgress)
}

// CheckConnection verifies if a connection can be established and closes it
// immediately; it returns the protocol info that was used.
func CheckConnection(address string, port string, protocol string, useTLS bool, tlsVerify bool, caCertPath string, clientCertPath string, clientKeyPath string) (string, error) {
	conn, err := DialConfig{
		Address:        address,
		Port:           port,
		Protocol:       protocol,
		UseTLS:         useTLS,
		TLSVerify:      tlsVerify,
		CACertPath:     caCertPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
	}.Dial()
	if err != nil {
		return "", err
	}
	_ = conn.Close()

	protocolInfo := protocol
	if useTLS {
		protocolInfo = fmt.Sprintf("%s+TLS", protocol)
	}
	return protocolInfo, nil
}

// sendTCPMessages sends TCP messages with proper framing per RFC 6587
func sendTCPMessages(conn net.Conn, config Config, onProgress SendFunc) Response {
	response := Response{
		SentMessages: []string{},
		Errors:       []string{},
	}

	framer := framing.NewFramer(framing.Config{
		Method:           config.Method,
		ValidateUTF8:     true,
		MaxMessageLength: 0,
	})

	totalMessages := len(config.Messages)
	for i, message := range config.Messages {
		syslogMsg, err := BuildMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		framedMsg, err := framer.Frame(syslogMsg)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error framing message: %v", err))
			continue
		}

		if err := conn.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
			// ignore deadline errors; write will surface them
			_ = err
		}

		if err := WriteAll(conn, framedMsg); err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}

		if onProgress != nil {
			onProgress(i+1, totalMessages, float64(i+1)/float64(totalMessages)*100)
		}
	}

	return response
}

// sendUDPMessages sends UDP messages (one message per packet)
func sendUDPMessages(conn net.Conn, config Config, onProgress SendFunc) Response {
	response := Response{
		SentMessages: []string{},
		Errors:       []string{},
	}

	totalMessages := len(config.Messages)
	for i, message := range config.Messages {
		syslogMsg, err := BuildMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		if err := conn.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
			_ = err
		}

		if err := WriteAll(conn, []byte(syslogMsg)); err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}

		if onProgress != nil {
			onProgress(i+1, totalMessages, float64(i+1)/float64(totalMessages)*100)
		}
	}

	return response
}

// ================================================================================
// MESSAGE FORMATTING
// ================================================================================

// BuildMessage constructs a syslog message per RFC 5424, RFC 3164 or the
// raw-pri passthrough mode.
func BuildMessage(config Config, message string) (string, error) {
	priority := config.Facility*8 + config.Severity

	if priority > 191 {
		return "", fmt.Errorf("invalid priority %d (facility=%d, severity=%d)", priority, config.Facility, config.Severity)
	}

	switch ResolveMessageFormat(config.MessageFormat, config.UseRFC5424) {
	case MessageFormatRawPRI:
		return BuildRawPRI(priority, message), nil
	case MessageFormatRFC3164:
		return buildRFC3164Message(priority, config, message), nil
	default:
		return buildRFC5424Message(priority, config, message), nil
	}
}

// BuildRawPRI emits a raw-pri message: the PRI value followed by the message
// verbatim, with no RFC 5424 header around it.
func BuildRawPRI(priority uint8, message string) string {
	return fmt.Sprintf("<%d>%s", priority, message)
}

func buildRFC5424Message(priority uint8, config Config, message string) string {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")

	hostname := config.Hostname
	if hostname == "" {
		hostname = "-"
	}

	appname := config.Appname
	if appname == "" {
		appname = "-"
	}

	return fmt.Sprintf("<%d>1 %s %s %s - - - %s",
		priority, timestamp, hostname, appname, message)
}

func buildRFC3164Message(priority uint8, config Config, message string) string {
	timestamp := time.Now().Format("Jan  2 15:04:05")

	hostname := config.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
		if hostname == "" {
			hostname = "localhost"
		}
	}

	appname := config.Appname
	if appname == "" {
		appname = "app"
	}

	return fmt.Sprintf("<%d>%s %s %s: %s",
		priority, timestamp, hostname, appname, message)
}

// WriteAll ensures all bytes are written (handles partial writes)
func WriteAll(w io.Writer, data []byte) error {
	totalWritten := 0
	dataLen := len(data)

	for totalWritten < dataLen {
		n, err := w.Write(data[totalWritten:])
		if err != nil {
			return fmt.Errorf("write failed after %d/%d bytes: %w", totalWritten, dataLen, err)
		}
		totalWritten += n
	}

	return nil
}

// ValidateConfig validates and normalizes configuration
func ValidateConfig(config *Config) error {
	if config.Address == "" {
		return fmt.Errorf("address is required")
	}
	if config.Port == "" {
		return fmt.Errorf("port is required")
	}

	if config.Facility > 23 {
		return fmt.Errorf("facility must be 0-23 (got %d)", config.Facility)
	}
	if config.Severity > 7 {
		return fmt.Errorf("severity must be 0-7 (got %d)", config.Severity)
	}

	if config.Method == "" {
		if config.Protocol == "tcp" {
			config.Method = framing.RecommendedFramingMethod()
		}
	}

	if config.Method != "" && !framing.IsValidFramingMethod(config.Method) {
		return fmt.Errorf("invalid framing method '%s'", config.Method)
	}

	if config.MessageFormat != "" && !IsValidMessageFormat(config.MessageFormat) {
		return fmt.Errorf("invalid message format '%s' (expected '%s', '%s' or '%s')",
			config.MessageFormat, MessageFormatRFC5424, MessageFormatRFC3164, MessageFormatRawPRI)
	}

	if config.Hostname == "" {
		hostname, err := os.Hostname()
		if err == nil && hostname != "" {
			config.Hostname = hostname
		} else {
			config.Hostname = "localhost"
		}
	}

	if config.Appname == "" {
		config.Appname = "sendlog"
	}

	return nil
}

// ================================================================================
// DIALING
// ================================================================================

// DialConfig is the subset of Config needed to establish a connection
type DialConfig struct {
	Address        string
	Port           string
	Protocol       string
	UseTLS         bool
	TLSVerify      bool
	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string
}

// Dial opens a connection based on protocol configuration
func Dial(config Config) (net.Conn, error) {
	return DialConfig{
		Address:        config.Address,
		Port:           config.Port,
		Protocol:       config.Protocol,
		UseTLS:         config.UseTLS,
		TLSVerify:      config.TLSVerify,
		CACertPath:     config.CACertPath,
		ClientCertPath: config.ClientCertPath,
		ClientKeyPath:  config.ClientKeyPath,
	}.Dial()
}

// Dial opens a connection using the DialConfig values
func (d DialConfig) Dial() (net.Conn, error) {
	if d.UseTLS && d.Protocol == "udp" {
		return nil, fmt.Errorf("TLS is not supported over UDP protocol")
	}

	if d.Protocol == "tcp" && d.UseTLS {
		return dialTLS(d)
	}

	fullAddress := net.JoinHostPort(d.Address, d.Port)
	dialer := &net.Dialer{
		Timeout: ConnectionTimeout,
	}

	conn, err := dialer.Dial(d.Protocol, fullAddress)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return conn, nil
}

// dialTLS establishes a TLS connection to the remote server
func dialTLS(d DialConfig) (net.Conn, error) {
	fullAddress := net.JoinHostPort(d.Address, d.Port)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !d.TLSVerify,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	if d.TLSVerify {
		tlsConfig.ServerName = d.Address

		if d.CACertPath != "" {
			rootCAs, err := loadCACertificate(d.CACertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA certificate: %w", err)
			}
			tlsConfig.RootCAs = rootCAs
		}
	}

	if d.ClientCertPath != "" && d.ClientKeyPath != "" {
		clientCert, err := loadClientCertificate(d.ClientCertPath, d.ClientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	} else if d.ClientCertPath != "" || d.ClientKeyPath != "" {
		return nil, fmt.Errorf("both client certificate and key are required for mTLS")
	}

	dialer := &net.Dialer{
		Timeout: ConnectionTimeout,
	}

	tlsConn, err := tls.DialWithDialer(dialer, "tcp", fullAddress, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}

	if err := tlsConn.Handshake(); err != nil {
		_ = tlsConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	return tlsConn, nil
}

// loadCACertificate loads a CA certificate from PEM file
func loadCACertificate(certPath string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
	}

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to parse CA certificate (invalid PEM format)")
	}

	return rootCAs, nil
}

// loadClientCertificate loads client certificate and private key for mTLS
func loadClientCertificate(certPath, keyPath string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load client certificate/key pair: %w", err)
	}
	return cert, nil
}
