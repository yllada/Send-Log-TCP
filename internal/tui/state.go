package tui

import (
	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/stress"
	"Sendlog-Syslog/internal/syslog"
)

// AppState is the single source of truth shared by every screen. Editing one
// screen updates the fields other screens read, so the form stays coherent as
// the user navigates.
type AppState struct {
	Address        string `json:"address"`
	Port           string `json:"port"`
	Protocol       string `json:"protocol"`
	FramingMethod  string `json:"framingMethod"`
	UseTLS         bool   `json:"useTls"`
	TLSVerify      bool   `json:"tlsVerify"`
	CACertPath     string `json:"caCertPath,omitempty"`
	ClientCertPath string `json:"clientCertPath,omitempty"`
	ClientKeyPath  string `json:"clientKeyPath,omitempty"`

	Format   string   `json:"format"`
	Facility uint8    `json:"facility"`
	Severity uint8    `json:"severity"`
	Hostname string   `json:"hostname"`
	Appname  string   `json:"appname"`
	Raw      bool     `json:"raw"`
	Messages []string `json:"messages"`

	StressMessage    string `json:"stressMessage"`
	StressMessagesPS int    `json:"stressMessagesPerSec"`
	StressDuration   int    `json:"stressDuration"`
	StressMaxMsgs    int    `json:"stressMaxMessages"`
}

func DefaultState() *AppState {
	return &AppState{
		Port:             "514",
		Protocol:         "tcp",
		FramingMethod:    "octet-counting",
		Format:           "rfc5424",
		Facility:         13,
		Severity:         6,
		Messages:         []string{},
		StressMessage:    "stress test message",
		StressMessagesPS: 10,
		StressDuration:   30,
	}
}

// ToSyslogConfig projects the shared state into the domain Config for sending.
func (s *AppState) ToSyslogConfig() syslog.Config {
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
		Hostname:       s.Hostname,
		Appname:        s.Appname,
		MessageFormat:  format,
		UseTLS:         s.UseTLS,
		TLSVerify:      s.TLSVerify,
		CACertPath:     s.CACertPath,
		ClientCertPath: s.ClientCertPath,
		ClientKeyPath:  s.ClientKeyPath,
	}
}

// ToStressConfig projects the shared state into the stress engine config.
func (s *AppState) ToStressConfig() stress.ContinuousSendConfig {
	return stress.ContinuousSendConfig{
		Address:        s.Address,
		Port:           s.Port,
		Protocol:       s.Protocol,
		Message:        s.StressMessage,
		Method:         framing.Method(s.FramingMethod),
		Facility:       s.Facility,
		Severity:       s.Severity,
		Hostname:       s.Hostname,
		Appname:        s.Appname,
		MessageFormat:  syslog.MessageFormat(s.Format),
		UseTLS:         s.UseTLS,
		TLSVerify:      s.TLSVerify,
		CACertPath:     s.CACertPath,
		ClientCertPath: s.ClientCertPath,
		ClientKeyPath:  s.ClientKeyPath,
		Duration:       s.StressDuration,
		MessagesPerSec: s.StressMessagesPS,
		MaxMessages:    s.StressMaxMsgs,
	}
}
