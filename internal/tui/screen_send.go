package tui

import (
	"fmt"
	"strings"

	"Sendlog-Syslog/internal/syslog"

	tea "github.com/charmbracelet/bubbletea"
)

type sendScreen struct {
	state        *AppState
	cursor       int
	sending      bool
	current      int
	total        int
	percent      float64
	sent         []string
	errors       []string
	sendMessages func(syslog.Config, syslog.SendFunc) syslog.Response
}

type sendProgressMsg struct {
	current int
	total   int
	percent float64
}

type sendResultMsg struct {
	resp syslog.Response
}

func NewSendScreen(state *AppState) tea.Model {
	return &sendScreen{
		state:        state,
		sendMessages: syslog.SendSyslogMessages,
	}
}

func (s *sendScreen) Init() tea.Cmd { return nil }

func (s *sendScreen) syncState() {}

func (s *sendScreen) itemCount() int {
	if len(s.state.Messages) == 0 {
		return 0
	}
	return len(s.state.Messages) + 1
}

func (s *sendScreen) isSendAction() bool {
	return len(s.state.Messages) > 0 && s.cursor == len(s.state.Messages)
}

func (s *sendScreen) moveUp() {
	if s.itemCount() == 0 {
		return
	}
	if s.cursor > 0 {
		s.cursor--
	}
}

func (s *sendScreen) moveDown() {
	if s.itemCount() == 0 {
		return
	}
	if s.cursor < s.itemCount()-1 {
		s.cursor++
	}
}

func (s *sendScreen) startSend() tea.Cmd {
	if !s.isSendAction() || s.sending {
		return nil
	}
	cfg := s.state.ToSyslogConfig()
	if err := syslog.ValidateConfig(&cfg); err != nil {
		s.errors = []string{err.Error()}
		return nil
	}
	s.sending = true
	s.sent = nil
	s.errors = nil
	s.current = 0
	s.total = len(cfg.Messages)
	s.percent = 0
	fn := s.sendMessages
	return func() tea.Msg {
		resp := fn(cfg, func(current, total int, percent float64) {
			s.current = current
			s.total = total
			s.percent = percent
		})
		return sendResultMsg{resp: resp}
	}
}

func (s *sendScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return s, nil
	case sendProgressMsg:
		s.current = msg.current
		s.total = msg.total
		s.percent = msg.percent
		return s, nil
	case sendResultMsg:
		s.sending = false
		s.sent = msg.resp.SentMessages
		s.errors = msg.resp.Errors
		return s, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if s.isSendAction() {
				return s, s.startSend()
			}
		case tea.KeyRunes:
			if len(msg.Runes) != 1 {
				return s, nil
			}
			switch msg.Runes[0] {
			case 's':
				if s.isSendAction() {
					return s, s.startSend()
				}
			case 'k':
				s.moveUp()
			case 'j':
				s.moveDown()
			}
		case tea.KeyUp:
			s.moveUp()
		case tea.KeyDown:
			s.moveDown()
		}
	}
	return s, nil
}

func (s *sendScreen) targetLine() string {
	if s.state.Address == "" {
		return "target: (set address in Connection)"
	}
	protocol := s.state.Protocol
	if s.state.UseTLS {
		protocol += "+TLS"
	}
	return fmt.Sprintf("target: %s:%s/%s", s.state.Address, s.state.Port, protocol)
}

func (s *sendScreen) View() string {
	var b strings.Builder
	b.WriteString(labelStyle(s.targetLine()))
	b.WriteString("\n")

	if len(s.state.Messages) == 0 {
		b.WriteString("\n")
		b.WriteString(dim.Render("no messages — add some in Messages or Import File first"))
	} else {
		b.WriteString("\n")
		b.WriteString(labelStyle(fmt.Sprintf("%d message(s)", len(s.state.Messages))))
		b.WriteString("\n")
		for i, msg := range s.state.Messages {
			cursor := " "
			if i == s.cursor {
				cursor = ">"
			}
			b.WriteString(labelStyle(fmt.Sprintf(" %s %2d ", cursor, i+1)))
			b.WriteString(msg)
			b.WriteString("\n")
		}
		cursor := " "
		if s.isSendAction() {
			cursor = ">"
		}
		b.WriteString(labelStyle(fmt.Sprintf(" %s     ", cursor)))
		if s.sending {
			b.WriteString(dim.Render("sending..."))
		} else {
			b.WriteString(accent.Render("send"))
		}
	}

	if s.sending {
		b.WriteString("\n")
		b.WriteString(okText(fmt.Sprintf("sending %d/%d (%.0f%%)", s.current, s.total, s.percent)))
	}

	if s.sent != nil {
		b.WriteString("\n")
		b.WriteString(okText(fmt.Sprintf("sent (%d):", len(s.sent))))
		b.WriteString("\n")
		for _, msg := range s.sent {
			b.WriteString("  " + msg + "\n")
		}
	}
	if len(s.errors) > 0 && !s.sending {
		if s.sent != nil {
			b.WriteString("\n")
		}
		b.WriteString(errText(fmt.Sprintf("failed (%d):", len(s.errors))))
		b.WriteString("\n")
		for _, e := range s.errors {
			b.WriteString("  " + e + "\n")
		}
	}
	if s.sent != nil {
		b.WriteString("\n")
		b.WriteString(accent.Render(fmt.Sprintf("summary: %d sent / %d failed", len(s.sent), len(s.errors))))
	}

	b.WriteString("\n\n")
	help := "↑/↓ move · s/enter send · esc back"
	if len(s.state.Messages) == 0 {
		help = "esc back"
	}
	b.WriteString(keyHelp(help))
	return screenBox("Send", b.String())
}
