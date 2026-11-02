package tui

import (
	"fmt"
	"strings"

	"Sendlog-Syslog/internal/syslog"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type connectionScreen struct {
	state      *AppState
	cursor     int
	address    textinput.Model
	port       textinput.Model
	caCert     textinput.Model
	clientCert textinput.Model
	clientKey  textinput.Model
	protocol   *OptionSelector
	framing    *OptionSelector
	tls        *OptionSelector
	validate   *OptionSelector
	status     string
	errMsg     string
}

// NewConnectionScreen builds the screen for editing the target server and transport settings.
func NewConnectionScreen(state *AppState) tea.Model {
	address := textinput.New()
	port := textinput.New()
	caCert := textinput.New()
	clientCert := textinput.New()
	clientKey := textinput.New()
	address.Width = 40
	address.Placeholder = "127.0.0.1"
	port.Width = 8
	port.Placeholder = "514"
	caCert.Placeholder = "/etc/ssl/certs/ca.pem"
	clientCert.Placeholder = "/path/to/client.crt"
	clientKey.Placeholder = "/path/to/client.key"
	m := &connectionScreen{
		state:      state,
		cursor:     0,
		address:    address,
		port:       port,
		caCert:     caCert,
		clientCert: clientCert,
		clientKey:  clientKey,
		protocol:   NewOptionSelector([]Option{{Value: "tcp", Label: "TCP"}, {Value: "udp", Label: "UDP"}}),
		framing:    NewOptionSelector([]Option{{Value: "octet-counting", Label: "octet-counting", Helper: "RFC 6587, recommended"}, {Value: "non-transparent", Label: "non-transparent", Helper: "newline-delimited"}}),
		tls:        NewOptionSelector([]Option{{Value: "off", Label: "off"}, {Value: "on", Label: "on"}}),
		validate:   NewOptionSelector([]Option{{Value: "verify", Label: "verify certs"}, {Value: "skip", Label: "skip verify"}}),
	}
	return m
}

func (s *connectionScreen) Init() tea.Cmd {
	return s.address.Focus()
}

func (s *connectionScreen) inputAt(i int) *textinput.Model {
	switch i {
	case 0:
		return &s.address
	case 1:
		return &s.port
	case 2:
		return &s.caCert
	case 3:
		return &s.clientCert
	case 4:
		return &s.clientKey
	default:
		return nil
	}
}

func (s *connectionScreen) blur() {
	for i := 0; i <= 4; i++ {
		s.inputAt(i).Blur()
	}
}

func (s *connectionScreen) syncState() {
	s.state.Address = s.address.Value()
	s.state.Port = s.port.Value()
	s.state.Protocol = s.protocol.Value()
	s.state.FramingMethod = s.framing.Value()
	s.state.UseTLS = s.tls.Value() == "on"
	s.state.TLSVerify = s.validate.Value() == "verify"
	s.state.CACertPath = s.caCert.Value()
	s.state.ClientCertPath = s.clientCert.Value()
	s.state.ClientKeyPath = s.clientKey.Value()
}

func (s *connectionScreen) testConnection() tea.Cmd {
	s.syncState()
	cfg := s.state.ToSyslogConfig()
	return func() tea.Msg {
		if err := syslog.ValidateConfig(&cfg); err != nil {
			return connectionResult{ok: false, err: "validation: " + err.Error()}
		}
		conn, err := syslog.Dial(cfg)
		if err != nil {
			return connectionResult{ok: false, err: err.Error()}
		}
		_ = conn.Close()
		return connectionResult{ok: true, msg: fmt.Sprintf("connected to %s via %s", cfg.Address, cfg.Protocol)}
	}
}

type connectionResult struct {
	ok  bool
	msg string
	err string
}

func (s *connectionScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return s, nil
	case connectionResult:
		if msg.ok {
			s.status = msg.msg
			s.errMsg = ""
		} else {
			s.errMsg = msg.err
			s.status = ""
		}
		return s, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyEnter:
			s.cursor++
			if s.cursor > 8 {
				s.cursor = 0
			}
			s.blur()
			if in := s.inputAt(s.cursor); in != nil {
				return s, in.Focus()
			}
			return s, nil
		case tea.KeyShiftTab:
			s.cursor--
			if s.cursor < 0 {
				s.cursor = 8
			}
			s.blur()
			if in := s.inputAt(s.cursor); in != nil {
				return s, in.Focus()
			}
			return s, nil
		case tea.KeyRunes, tea.KeyBackspace, tea.KeySpace:
			if in := s.inputAt(s.cursor); in != nil {
				var cmd tea.Cmd
				*in, cmd = in.Update(msg)
				return s, cmd
			}
			if s.cursor >= 5 && len(msg.Runes) == 1 && msg.Runes[0] == 't' {
				return s, s.testConnection()
			}
			if s.cursor == 5 && len(msg.Runes) == 1 {
				switch msg.Runes[0] {
				case 'l':
					s.protocol.Toggle()
					return s, nil
				case 'h':
					s.protocol.Up()
					return s, nil
				case 'j':
					s.protocol.Down()
					return s, nil
				}
			}
			if s.cursor == 6 && s.state.Protocol == "tcp" && len(msg.Runes) == 1 && msg.Runes[0] == 'l' {
				s.framing.Toggle()
				return s, nil
			}
			if s.cursor == 7 && len(msg.Runes) == 1 && msg.Runes[0] == 'l' {
				s.tls.Toggle()
				return s, nil
			}
			if s.cursor == 8 && len(msg.Runes) == 1 && msg.Runes[0] == 'l' {
				s.validate.Toggle()
				return s, nil
			}
		}
	}
	return s, nil
}

func (s *connectionScreen) View() string {
	if s.inputAt(s.cursor) != nil {
		s.inputAt(s.cursor).Cursor.Blink = false
	}
	var b strings.Builder
	b.WriteString(titleStyle("Connection"))
	b.WriteString("\n\n")
	b.WriteString(labelStyle("Host:     "))
	b.WriteString(s.address.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Port:     "))
	b.WriteString(s.port.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Protocol: "))
	b.WriteString(s.protocol.View())
	b.WriteString("\n")
	if s.state.Protocol == "tcp" {
		b.WriteString(labelStyle("Framing:  "))
		b.WriteString(s.framing.View())
		b.WriteString("\n")
	}
	b.WriteString(labelStyle("TLS:      "))
	b.WriteString(s.tls.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Verify:   "))
	b.WriteString(s.validate.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("CA cert:  "))
	b.WriteString(s.caCert.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Cert:     "))
	b.WriteString(s.clientCert.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Key:      "))
	b.WriteString(s.clientKey.View())
	b.WriteString("\n")
	if s.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(errText(s.errMsg))
	} else if s.status != "" {
		b.WriteString("\n")
		b.WriteString(okText(s.status))
	}
	b.WriteString("\n\n")
	b.WriteString(keyHelp("tab/enter next · backtab prev · l toggle · t test · esc back"))
	return screenBox("Connection", b.String())
}
