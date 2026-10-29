package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/syslog"
)

// ================================================================================
// SYSLOG SERVICE - GUI-facing wrapper around the internal/syslog engine
// The message building, framing and sending logic lives in internal/syslog;
// this service maps the Wails runtime context and events onto it.
// ================================================================================

// SyslogConfig aliases the internal config so the frontend binding shape is unchanged
type SyslogConfig = syslog.SyslogConfig
type SyslogResponse = syslog.SyslogResponse

// FramingMethod aliases the internal framing method
type FramingMethod = framing.FramingMethod

const (
	// OctetCounting implements octet counting method (RFC 6587 Section 3.4.1)
	OctetCounting = framing.OctetCounting

	// NonTransparent implements non-transparent framing (RFC 6587 Section 3.4.2)
	NonTransparent = framing.NonTransparent
)

// MessageFormat aliases the internal message format
type MessageFormat = syslog.MessageFormat

const (
	// MessageFormatRFC5424 builds a full RFC 5424 header
	MessageFormatRFC5424 = syslog.MessageFormatRFC5424

	// MessageFormatRFC3164 builds a legacy BSD syslog header
	MessageFormatRFC3164 = syslog.MessageFormatRFC3164

	// MessageFormatRawPRI emits only "<PRI>" followed by the message bytes
	MessageFormatRawPRI = syslog.MessageFormatRawPRI
)

// IsValidMessageFormat reports whether the message format is recognized
func IsValidMessageFormat(format MessageFormat) bool {
	return syslog.IsValidMessageFormat(format)
}

// SyslogService handles syslog message sending operations for the GUI
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
	// Emit start event
	runtime.EventsEmit(s.ctx, "syslog:sending", map[string]interface{}{
		"total": len(config.Messages),
	})

	response := syslog.SendSyslogMessages(config, func(current, total int, percent float64) {
		runtime.EventsEmit(s.ctx, "syslog:progress", map[string]interface{}{
			"current": current,
			"total":   total,
			"percent": percent,
		})
	})

	// Emit completion event
	runtime.EventsEmit(s.ctx, "syslog:complete", map[string]interface{}{
		"sent":   len(response.SentMessages),
		"errors": len(response.Errors),
	})

	return response
}
