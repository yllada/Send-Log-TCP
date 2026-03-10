package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ================================================================================
// SYSLOG SERVICE - Handles syslog message formatting and sending
// Follows Single Responsibility Principle: manages message formatting and transmission
// ================================================================================

// FramingMethod specifies the framing method for TCP per RFC 6587
type FramingMethod string

const (
	// OctetCounting implements octet counting method (RFC 6587 Section 3.4.1)
	OctetCounting FramingMethod = "octet-counting"

	// NonTransparent implements non-transparent framing (RFC 6587 Section 3.4.2)
	NonTransparent FramingMethod = "non-transparent"
)

// SyslogConfig holds the configuration for sending syslog messages
type SyslogConfig struct {
	Address        string        `json:"Address"`
	Port           string        `json:"Port"`
	Protocol       string        `json:"Protocol"`
	Messages       []string      `json:"Messages"`
	FramingMethod  FramingMethod `json:"FramingMethod"`
	Facility       uint8         `json:"Facility"`
	Severity       uint8         `json:"Severity"`
	Hostname       string        `json:"Hostname"`
	Appname        string        `json:"Appname"`
	UseRFC5424     bool          `json:"UseRFC5424"`
	UseTLS         bool          `json:"UseTLS"`
	TLSVerify      bool          `json:"TLSVerify"`
	CACertPath     string        `json:"CACertPath"`
	ClientCertPath string        `json:"ClientCertPath"`
	ClientKeyPath  string        `json:"ClientKeyPath"`
}

// SyslogResponse contains the result of send operations
type SyslogResponse struct {
	SentMessages []string `json:"sentMessages"`
	Errors       []string `json:"errors"`
}

// SyslogService handles syslog message sending operations
type SyslogService struct {
	ctx context.Context
}

// NewSyslogService creates a new SyslogService instance
func NewSyslogService() *SyslogService {
	return &SyslogService{}
}

// SetContext sets the Wails runtime context
func (s *SyslogService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SendSyslogMessages sends syslog messages and returns the result
func (s *SyslogService) SendSyslogMessages(config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Invalid configuration: %v", err))
		return response
	}

	// Emit start event
	runtime.EventsEmit(s.ctx, "syslog:sending", map[string]interface{}{
		"total": len(config.Messages),
	})

	fullAddress := net.JoinHostPort(config.Address, config.Port)

	// Establish connection
	conn, err := dialConnection(config.Address, config.Port, config.Protocol, config.UseTLS, config.TLSVerify, config.CACertPath, config.ClientCertPath, config.ClientKeyPath)
	if err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Error connecting to %s: %v", fullAddress, err))
		return response
	}
	defer conn.Close()

	// Send messages based on protocol
	if config.Protocol == "tcp" {
		return s.sendTCPMessages(conn, config)
	}
	return s.sendUDPMessages(conn, config)
}

// sendTCPMessages sends TCP messages with proper framing per RFC 6587
func (s *SyslogService) sendTCPMessages(conn net.Conn, config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	// Create framer with appropriate configuration
	framer := NewFramer(FramingConfig{
		Method:           config.FramingMethod,
		ValidateUTF8:     true,
		MaxMessageLength: 0,
	})

	totalMessages := len(config.Messages)
	for i, message := range config.Messages {
		// Build syslog message
		syslogMsg, err := buildSyslogMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		// Apply TCP framing
		framedMsg, err := framer.Frame(syslogMsg)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error framing message: %v", err))
			continue
		}

		// Set write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to set write deadline: %v", err))
		}

		// Send with robust handling
		if err := writeAll(conn, framedMsg); err != nil {
			runtime.LogError(s.ctx, fmt.Sprintf("Error sending message: %v", err))
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			runtime.LogDebug(s.ctx, fmt.Sprintf("Sent message: %s", syslogMsg))
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}

		// Emit progress event
		runtime.EventsEmit(s.ctx, "syslog:progress", map[string]interface{}{
			"current": i + 1,
			"total":   totalMessages,
			"percent": float64(i+1) / float64(totalMessages) * 100,
		})
	}

	// Emit completion event
	runtime.EventsEmit(s.ctx, "syslog:complete", map[string]interface{}{
		"sent":   len(response.SentMessages),
		"errors": len(response.Errors),
	})

	return response
}

// sendUDPMessages sends UDP messages (one message per packet)
func (s *SyslogService) sendUDPMessages(conn net.Conn, config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	totalMessages := len(config.Messages)
	for i, message := range config.Messages {
		syslogMsg, err := buildSyslogMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to set write deadline: %v", err))
		}

		if err := writeAll(conn, []byte(syslogMsg)); err != nil {
			runtime.LogError(s.ctx, fmt.Sprintf("Error sending message: %v", err))
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			runtime.LogDebug(s.ctx, fmt.Sprintf("Sent message: %s", syslogMsg))
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}

		runtime.EventsEmit(s.ctx, "syslog:progress", map[string]interface{}{
			"current": i + 1,
			"total":   totalMessages,
			"percent": float64(i+1) / float64(totalMessages) * 100,
		})
	}

	runtime.EventsEmit(s.ctx, "syslog:complete", map[string]interface{}{
		"sent":   len(response.SentMessages),
		"errors": len(response.Errors),
	})

	return response
}

// ================================================================================
// SYSLOG MESSAGE FORMATTING HELPERS
// ================================================================================

// buildSyslogMessage constructs a valid syslog message per RFC 5424 or RFC 3164
func buildSyslogMessage(config SyslogConfig, message string) (string, error) {
	priority := config.Facility*8 + config.Severity

	if priority > 191 {
		return "", fmt.Errorf("invalid priority %d (facility=%d, severity=%d)", priority, config.Facility, config.Severity)
	}

	if config.UseRFC5424 {
		return buildRFC5424Message(priority, config, message), nil
	}
	return buildRFC3164Message(priority, config, message), nil
}

// buildRFC5424Message constructs message per RFC 5424
func buildRFC5424Message(priority uint8, config SyslogConfig, message string) string {
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

// buildRFC3164Message constructs message per RFC 3164 (BSD syslog)
func buildRFC3164Message(priority uint8, config SyslogConfig, message string) string {
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

// writeAll ensures all bytes are written (handles partial writes)
func writeAll(w io.Writer, data []byte) error {
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

// validateConfig validates and normalizes configuration
func validateConfig(config *SyslogConfig) error {
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

	if config.FramingMethod == "" {
		if config.Protocol == "tcp" {
			config.FramingMethod = RecommendedFramingMethod()
		}
	}

	if config.FramingMethod != "" && !IsValidFramingMethod(config.FramingMethod) {
		return fmt.Errorf("invalid framing method '%s'", config.FramingMethod)
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
