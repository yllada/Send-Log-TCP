//go:build linux && cgo

package gui

import (
	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/stress"
	"Sendlog-Syslog/internal/syslog"
)

// State mirrors the TUI application state; the GUI presents the same domain
// values without importing the TUI layer.
type State struct {
	Address        string
	Port           string
	Protocol       string
	FramingMethod  string
	UseTLS         bool
	TLSVerify      bool
	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string

	Format   string
	Facility uint8
	Severity uint8
	Raw      bool
	Messages []string

	StressMessage  string
	StressRate     int
	StressDuration int
	StressMax      int
	RandomizeData  bool
}

// DefaultState returns the GUI state seeded with the application defaults.
func DefaultState() *State {
	return &State{
		Port:           "514",
		Protocol:       "tcp",
		FramingMethod:  "octet-counting",
		Format:         "rfc5424",
		Facility:       13,
		Severity:       6,
		Messages:       []string{},
		StressMessage:  "stress test message",
		StressRate:     10,
		StressDuration: 0,
		StressMax:      0,
	}
}

// ToSyslogConfig projects the GUI state into the domain send configuration.
func (s *State) ToSyslogConfig() syslog.Config {
	format := syslog.MessageFormat(s.Format)
	if s.Raw {
		format = syslog.MessageFormatRawPRI
	}
	return syslog.Config{
		Address:        s.Address,
		Port:           s.Port,
		Protocol:       s.Protocol,
		Messages:       s.Messages,
		Method:         framing.Method(s.FramingMethod),
		Facility:       s.Facility,
		Severity:       s.Severity,
		MessageFormat:  format,
		UseTLS:         s.UseTLS,
		TLSVerify:      s.TLSVerify,
		CACertPath:     s.CACertPath,
		ClientCertPath: s.ClientCertPath,
		ClientKeyPath:  s.ClientKeyPath,
	}
}

// ToStressConfig projects the GUI state into the continuous send configuration.
func (s *State) ToStressConfig() stress.ContinuousSendConfig {
	return stress.ContinuousSendConfig{
		Address:        s.Address,
		Port:           s.Port,
		Protocol:       s.Protocol,
		Message:        s.StressMessage,
		Method:         framing.Method(s.FramingMethod),
		Facility:       s.Facility,
		Severity:       s.Severity,
		MessageFormat:  syslog.MessageFormat(s.Format),
		UseTLS:         s.UseTLS,
		TLSVerify:      s.TLSVerify,
		CACertPath:     s.CACertPath,
		ClientCertPath: s.ClientCertPath,
		ClientKeyPath:  s.ClientKeyPath,
		Duration:       s.StressDuration,
		MessagesPerSec: s.StressRate,
		MaxMessages:    s.StressMax,
		RandomizeData:  s.RandomizeData,
	}
}
