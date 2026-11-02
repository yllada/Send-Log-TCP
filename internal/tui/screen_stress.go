package tui

import (
	"fmt"
	"strings"
	"time"

	"Sendlog-Syslog/internal/stress"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type stressRunner interface {
	Start(config stress.ContinuousSendConfig, onStats func(stress.ContinuousStats)) error
	Stop()
	IsContinuousRunning() bool
	GetContinuousStats() stress.ContinuousStats
}

type stressScreen struct {
	state     *AppState
	cursor    int
	message   textinput.Model
	rate      textinput.Model
	duration  textinput.Model
	maxInput  textinput.Model
	runner    stressRunner
	errMsg    string
	running   bool
	started   bool
	summary   string
	statsLine string
	failed    bool
}

// NewStressScreen builds the screen for running and monitoring the continuous send engine.
func NewStressScreen(state *AppState) tea.Model {
	message := textinput.New()
	message.Width = 60
	message.Placeholder = "message sent on every tick"
	rate := textinput.New()
	rate.Width = 8
	duration := textinput.New()
	duration.Width = 8
	maxInput := textinput.New()
	maxInput.Width = 8
	return &stressScreen{
		state:    state,
		message:  message,
		rate:     rate,
		duration: duration,
		maxInput: maxInput,
		runner:   stress.NewEngine(),
	}
}

func (s *stressScreen) Init() tea.Cmd { return s.message.Focus() }

func (s *stressScreen) inputAt(i int) *textinput.Model {
	switch i {
	case 0:
		return &s.message
	case 1:
		return &s.rate
	case 2:
		return &s.duration
	case 3:
		return &s.maxInput
	default:
		return nil
	}
}

func (s *stressScreen) blurAll() {
	for i := 0; i <= 3; i++ {
		s.inputAt(i).Blur()
	}
}

func (s *stressScreen) syncState() {
	if s.started {
		// Leaving the screen stops the engine so nothing keeps sending
		// from the background.
		s.runner.Stop()
		s.started = false
	}
	if s.message.Value() != "" {
		s.state.StressMessage = s.message.Value()
	}
	if v := atoiField(s.rate.Value()); v > 0 {
		s.state.StressMessagesPS = v
	}
	s.state.StressDuration = atoiField(s.duration.Value())
	s.state.StressMaxMsgs = atoiField(s.maxInput.Value())
}

func atoiField(v string) int {
	n := 0
	for _, r := range v {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func (s *stressScreen) isAction() bool { return s.cursor == 4 }

func (s *stressScreen) moveCursor(delta int) {
	s.blurAll()
	s.cursor = (s.cursor + delta + 5) % 5
	if in := s.inputAt(s.cursor); in != nil {
		in.Focus()
	}
}

func (s *stressScreen) start() tea.Cmd {
	s.syncState()
	msg := s.message.Value()
	if msg == "" && len(s.state.Messages) > 0 {
		msg = s.state.Messages[0]
	}
	if msg == "" {
		s.errMsg = "set a message first"
		return nil
	}
	rate := atoiField(s.rate.Value())
	if rate <= 0 {
		s.errMsg = "rate must be a positive number"
		return nil
	}
	cfg := s.state.ToStressConfig()
	cfg.Message = msg
	cfg.MessagesPerSec = rate
	cfg.Duration = atoiField(s.duration.Value())
	cfg.MaxMessages = atoiField(s.maxInput.Value())
	runner := s.runner
	return func() tea.Msg {
		if err := runner.Start(cfg, nil); err != nil {
			return stressStartFailedMsg{err: err.Error()}
		}
		return stressStartedMsg{}
	}
}

func (s *stressScreen) stop() tea.Cmd {
	runner := s.runner
	return func() tea.Msg {
		runner.Stop()
		return stressStatsMsg{stats: runner.GetContinuousStats()}
	}
}

func (s *stressScreen) tick() tea.Cmd {
	runner := s.runner
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return stressStatsMsg{stats: runner.GetContinuousStats()}
	})
}

type stressStartedMsg struct{}

type stressStartFailedMsg struct{ err string }

type stressStatsMsg struct{ stats stress.ContinuousStats }

func (s *stressScreen) finish(st stress.ContinuousStats) {
	s.started = false
	s.running = false
	s.statsLine = ""
	s.blurAll()
	if st.TotalSent == 0 && st.TotalErrors == 0 && st.ElapsedSeconds < 1 {
		s.failed = true
		s.summary = "no messages sent — the engine could not reach the server"
		return
	}
	s.summary = fmt.Sprintf("stopped after %.1fs — sent=%d errors=%d avg rate=%.1f/s",
		st.ElapsedSeconds, st.TotalSent, st.TotalErrors,
		float64(st.TotalSent)/max(st.ElapsedSeconds, 0.001))
}

func (s *stressScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return s, nil
	case stressStartedMsg:
		s.started = true
		s.running = true
		s.failed = false
		s.summary = ""
		s.errMsg = ""
		s.blurAll()
		return s, s.tick()
	case stressStartFailedMsg:
		s.errMsg = msg.err
		return s, nil
	case stressStatsMsg:
		if msg.stats.IsRunning {
			s.running = true
			s.statsLine = fmt.Sprintf("sent=%d err=%d rate=%.0f/s elapsed=%.0fs target=%d/s",
				msg.stats.TotalSent, msg.stats.TotalErrors, msg.stats.CurrentRate,
				msg.stats.ElapsedSeconds, msg.stats.TargetRate)
			return s, s.tick()
		}
		s.finish(msg.stats)
		return s, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			if s.started && s.isAction() {
				return s, s.stop()
			}
			s.moveCursor(1)
			return s, nil
		case tea.KeyShiftTab:
			s.moveCursor(-1)
			return s, nil
		case tea.KeyUp:
			s.moveCursor(-1)
			return s, nil
		case tea.KeyDown:
			s.moveCursor(1)
			return s, nil
		case tea.KeyEnter, tea.KeySpace:
			if s.isAction() {
				if s.started {
					return s, s.stop()
				}
				return s, s.start()
			}
			if !s.started {
				if in := s.inputAt(s.cursor); in != nil {
					var cmd tea.Cmd
					*in, cmd = in.Update(msg)
					return s, cmd
				}
			}
			return s, nil
		case tea.KeyRunes, tea.KeyBackspace:
			if s.started {
				if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 &&
					(msg.Runes[0] == 's' || msg.Runes[0] == 'q') {
					return s, s.stop()
				}
				return s, nil
			}
			if in := s.inputAt(s.cursor); in != nil {
				var cmd tea.Cmd
				*in, cmd = in.Update(msg)
				return s, cmd
			}
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 &&
				msg.Runes[0] == 's' && s.isAction() {
				return s, s.start()
			}
		}
	}
	return s, nil
}

func (s *stressScreen) targetLine() string {
	if s.state.Address == "" {
		return "target: (set address in Connection)"
	}
	protocol := s.state.Protocol
	if s.state.UseTLS {
		protocol += "+TLS"
	}
	return fmt.Sprintf("target: %s:%s/%s", s.state.Address, s.state.Port, protocol)
}

func (s *stressScreen) View() string {
	if !s.started {
		if in := s.inputAt(s.cursor); in != nil {
			in.Cursor.Blink = false
		}
	}
	var b strings.Builder
	b.WriteString(labelStyle(s.targetLine()))
	b.WriteString("\n")

	if s.running {
		b.WriteString(okText("running: " + s.statsLine))
		b.WriteString("\n")
	} else if s.summary != "" {
		if s.failed {
			b.WriteString(errText(s.summary))
		} else {
			b.WriteString(okText(s.summary))
		}
		b.WriteString("\n")
	}
	if s.errMsg != "" {
		b.WriteString(errText(s.errMsg))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(labelStyle("Message:  "))
	b.WriteString(s.message.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Rate:     "))
	b.WriteString(s.rate.View())
	b.WriteString("  messages per second\n")
	b.WriteString(labelStyle("Duration: "))
	b.WriteString(s.duration.View())
	b.WriteString("  seconds (blank = until stopped)\n")
	b.WriteString(labelStyle("Max:      "))
	b.WriteString(s.maxInput.View())
	b.WriteString("  messages (blank = unlimited)\n")
	b.WriteString("\n")
	b.WriteString(labelStyle("         "))
	if s.started {
		b.WriteString(accent.Render("stop"))
	} else {
		b.WriteString(accent.Render("start"))
	}
	b.WriteString("\n\n")

	help := "tab move · enter/s start · esc back"
	if s.started {
		help = "q/s stop · tab move · esc back"
	}
	b.WriteString(keyHelp(help))
	return screenBox("Stress", b.String())
}
