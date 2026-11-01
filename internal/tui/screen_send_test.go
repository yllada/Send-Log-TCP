package tui

import (
	"strings"
	"testing"

	"Sendlog-Syslog/internal/syslog"

	tea "github.com/charmbracelet/bubbletea"
)

func sendState(messages []string) *AppState {
	state := DefaultState()
	state.Address = "127.0.0.1"
	state.Messages = messages
	return state
}

type fakeSender struct {
	called bool
	cfg    syslog.Config
	resp   syslog.Response
}

func (f *fakeSender) send(cfg syslog.Config, onProgress syslog.SendFunc) syslog.Response {
	f.called = true
	f.cfg = cfg
	for i := range cfg.Messages {
		onProgress(i+1, len(cfg.Messages), float64(i+1)/float64(len(cfg.Messages))*100)
	}
	return f.resp
}

func updateScreen(t *testing.T, s *sendScreen, msg tea.Msg) *sendScreen {
	t.Helper()
	next, _ := s.Update(msg)
	screen, ok := next.(*sendScreen)
	if !ok {
		t.Fatal("Update returned wrong model type")
	}
	return screen
}

func runCmd(t *testing.T, s *sendScreen, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		return nil
	}
	return cmd()
}

func TestSendScreenShowsListTargetCountAndAction(t *testing.T) {
	s := NewSendScreen(sendState([]string{"first msg", "second msg"})).(*sendScreen)

	view := s.View()
	for _, want := range []string{
		"target: 127.0.0.1:514/tcp",
		"2 message(s)",
		"first msg",
		"second msg",
		"send",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("View missing %q", want)
		}
	}
}

func TestSendMovesCursorAndFiresOnlyOnAction(t *testing.T) {
	fake := &fakeSender{resp: syslog.Response{SentMessages: []string{"s1"}}}
	s := NewSendScreen(sendState([]string{"a", "b"})).(*sendScreen)
	s.sendMessages = fake.send

	s = updateScreen(t, s, tea.KeyMsg{Type: tea.KeyUp})
	if _, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatal("enter on a message line must not start a send")
	}
	s = updateScreen(t, s, tea.KeyMsg{Type: tea.KeyDown})
	if _, cmd := s.Update(tea.KeyMsg{Type: tea.KeyDown}); cmd != nil {
		t.Fatal("down past the send action must not move the cursor further")
	}

	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		t.Fatal("cursor is already on the send action")
	}
}

func TestSendRunsFakeAndRendersResults(t *testing.T) {
	fake := &fakeSender{resp: syslog.Response{
		SentMessages: []string{"<106>1 first built", "<106>1 second built"},
		Errors:       []string{},
	}}
	s := NewSendScreen(sendState([]string{"first", "second"})).(*sendScreen)
	s.sendMessages = fake.send

	for i := 0; i < 3; i++ {
		s = updateScreen(t, s, tea.KeyMsg{Type: tea.KeyDown})
	}
	if !s.isSendAction() {
		t.Fatal("cursor should be on the send action")
	}

	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := runCmd(t, s, cmd)
	if msg == nil {
		t.Fatal("enter on the send action must return a cmd")
	}
	if !fake.called {
		t.Fatal("fake sender was not called")
	}
	if len(fake.cfg.Messages) != 2 || fake.cfg.Messages[0] != "first" || fake.cfg.Address != "127.0.0.1" || fake.cfg.Protocol != "tcp" {
		t.Fatalf("fake got wrong config: %+v", fake.cfg)
	}

	s = updateScreen(t, s, msg)
	view := s.View()
	for _, want := range []string{
		"<106>1 first built",
		"<106>1 second built",
		"summary: 2 sent / 0 failed",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("View missing %q", want)
		}
	}
	if s.sending {
		t.Fatal("send must be finished")
	}
}

func TestSendShowsErrorsInResults(t *testing.T) {
	fake := &fakeSender{resp: syslog.Response{
		SentMessages: []string{"<106>1 ok"},
		Errors:       []string{"Error sending message: write failed"},
	}}
	s := NewSendScreen(sendState([]string{"one", "two"})).(*sendScreen)
	s.sendMessages = fake.send

	for i := 0; i < 3; i++ {
		s = updateScreen(t, s, tea.KeyMsg{Type: tea.KeyDown})
	}
	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	s = updateScreen(t, s, runCmd(t, s, cmd))

	view := s.View()
	for _, want := range []string{
		"failed (1):",
		"Error sending message: write failed",
		"summary: 1 sent / 1 failed",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("View missing %q", want)
		}
	}
}

func TestSendShowsProgressWhileSending(t *testing.T) {
	fake := &fakeSender{resp: syslog.Response{SentMessages: []string{"ok"}}}
	s := NewSendScreen(sendState([]string{"only"})).(*sendScreen)
	s.sendMessages = fake.send

	s = updateScreen(t, s, tea.KeyMsg{Type: tea.KeyDown})
	s = updateScreen(t, s, tea.KeyMsg{Type: tea.KeyDown})
	if _, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}); cmd == nil {
		t.Fatal("s on the send action must return a cmd")
	}

	s = updateScreen(t, s, sendProgressMsg{current: 1, total: 1, percent: 50})
	view := s.View()
	if !strings.Contains(view, "sending 1/1") {
		t.Error("View missing live progress")
	}
}

func TestSendWithoutMessagesDoesNotFire(t *testing.T) {
	fake := &fakeSender{}
	s := NewSendScreen(sendState(nil)).(*sendScreen)
	s.sendMessages = fake.send

	view := s.View()
	if !strings.Contains(view, "no messages") {
		t.Error("View missing the compose/import hint")
	}
	if strings.Contains(view, "> ") {
		t.Error("no cursor should be shown without messages")
	}

	if _, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatal("enter with no messages must not return a send cmd")
	}
	if _, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}); cmd != nil {
		t.Fatal("s with no messages must not return a send cmd")
	}
	if fake.called {
		t.Fatal("sender must not be called with no messages")
	}
}

func TestSendValidationFailureShowsErrorAndSendsNothing(t *testing.T) {
	fake := &fakeSender{}
	state := sendState([]string{"m1"})
	state.Address = ""
	s := NewSendScreen(state).(*sendScreen)
	s.sendMessages = fake.send

	s = updateScreen(t, s, tea.KeyMsg{Type: tea.KeyDown})
	s = updateScreen(t, s, tea.KeyMsg{Type: tea.KeyDown})
	if _, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatal("invalid config must not return a send cmd")
	}
	if fake.called {
		t.Fatal("sender must not be called when validation fails")
	}

	view := s.View()
	if !strings.Contains(view, "address is required") {
		t.Error("View missing the validation error")
	}
	if strings.Contains(view, "summary:") {
		t.Error("no summary should be shown before any send")
	}
}
