package tui

import (
	"strings"
	"testing"

	"Sendlog-Syslog/internal/framing"
	"Sendlog-Syslog/internal/stress"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeRunner struct {
	started    bool
	startErr   error
	stopped    bool
	lastConfig stress.ContinuousSendConfig
	current    stress.ContinuousStats
	afterStop  stress.ContinuousStats
}

func (f *fakeRunner) Start(config stress.ContinuousSendConfig, _ func(stress.ContinuousStats)) error {
	if f.startErr != nil {
		return f.startErr
	}
	f.started = true
	f.lastConfig = config
	f.current = stress.ContinuousStats{
		IsRunning:  true,
		StartTime:  1,
		TargetRate: config.MessagesPerSec,
		Duration:   config.Duration,
	}
	return nil
}

func (f *fakeRunner) Stop() {
	if f.stopped {
		return
	}
	f.stopped = true
	f.current.IsRunning = false
	f.current.TotalSent = 42
	f.current.TotalErrors = 1
	f.current.ElapsedSeconds = 5
	f.current.CurrentRate = 8.4
	f.afterStop = f.current
}

func (f *fakeRunner) IsContinuousRunning() bool { return f.started && !f.stopped }

func (f *fakeRunner) GetContinuousStats() stress.ContinuousStats {
	if f.stopped {
		return f.afterStop
	}
	return f.current
}

func newStressScreenForTest(t *testing.T) (*stressScreen, *fakeRunner) {
	t.Helper()
	fake := &fakeRunner{}
	scr, ok := NewStressScreen(DefaultState()).(*stressScreen)
	if !ok {
		t.Fatal("NewStressScreen did not return *stressScreen")
	}
	scr.runner = fake
	return scr, fake
}

func stressInput(t *testing.T, scr *stressScreen, field int, text string) {
	t.Helper()
	scr.blurAll()
	scr.cursor = field
	scr.inputAt(field).Focus()
	for _, r := range text {
		m, _ := scr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		scr = m.(*stressScreen)
	}
}

func TestStressEmptyMessageShowsErrorAndDoesNotStart(t *testing.T) {
	scr, fake := newStressScreenForTest(t)
	scr.cursor = 4
	_, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		cmd()
	}
	if fake.started {
		t.Fatal("engine started with no message")
	}
	if !strings.Contains(scr.errMsg, "set a message first") {
		t.Fatalf("expected empty-message error, got %q", scr.errMsg)
	}
	if !strings.Contains(scr.View(), "set a message first") {
		t.Fatal("View does not show the empty-message error")
	}
}

func TestStressStartSendsConfigToEngine(t *testing.T) {
	scr, fake := newStressScreenForTest(t)
	stressInput(t, scr, 0, "hello")
	stressInput(t, scr, 1, "2")
	stressInput(t, scr, 2, "30")
	stressInput(t, scr, 3, "100")
	scr.cursor = 4
	_, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a start command")
	}
	cmd()

	if !fake.started {
		t.Fatal("engine not started")
	}
	if fake.lastConfig.Message != "hello" {
		t.Fatalf("message = %q", fake.lastConfig.Message)
	}
	if fake.lastConfig.MessagesPerSec != 2 {
		t.Fatalf("rate = %d", fake.lastConfig.MessagesPerSec)
	}
	if fake.lastConfig.Duration != 30 {
		t.Fatalf("duration = %d", fake.lastConfig.Duration)
	}
	if fake.lastConfig.MaxMessages != 100 {
		t.Fatalf("max = %d", fake.lastConfig.MaxMessages)
	}
}

func TestStressStatsUpdateRenderedTotals(t *testing.T) {
	scr, fake := newStressScreenForTest(t)
	stressInput(t, scr, 0, "hello")
	stressInput(t, scr, 1, "2")
	scr.cursor = 4
	_, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	cmd()
	fake.current.TotalSent = 50
	fake.current.TotalErrors = 3
	fake.current.CurrentRate = 2.0
	fake.current.ElapsedSeconds = 25
	m, _ := scr.Update(stressStatsMsg{stats: fake.current})
	scr = m.(*stressScreen)
	view := scr.View()
	for _, want := range []string{"sent=50", "err=3", "elapsed=25s", "target=2/s"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View missing %q:\n%s", want, view)
		}
	}
}

func TestStressStopCallsEngineAndShowsSummary(t *testing.T) {
	scr, fake := newStressScreenForTest(t)
	stressInput(t, scr, 0, "hello")
	stressInput(t, scr, 1, "2")
	scr.cursor = 4
	_, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a start command")
	}
	if msg := cmd().(stressStartedMsg); msg != (stressStartedMsg{}) {
		t.Fatal("expected stressStartedMsg")
	}
	_, _ = scr.Update(stressStartedMsg{})
	_, cmd = scr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected a stop command")
	}
	cmd()
	m, _ := scr.Update(stressStatsMsg{stats: fake.afterStop})
	scr = m.(*stressScreen)
	if !fake.stopped {
		t.Fatal("engine not stopped")
	}
	if scr.started {
		t.Fatal("screen still marked started")
	}
	view := scr.View()
	if !strings.Contains(view, "sent=42") || !strings.Contains(view, "errors=1") {
		t.Fatalf("View missing final summary:\n%s", view)
	}
}

func TestStressEngineFinishShowsSummaryAndAllowsRestart(t *testing.T) {
	scr, fake := newStressScreenForTest(t)
	stressInput(t, scr, 0, "hello")
	stressInput(t, scr, 1, "2")
	scr.cursor = 4
	_, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	cmd()
	done := fake.current
	done.IsRunning = false
	done.TotalSent = 10
	done.ElapsedSeconds = 5
	m, _ := scr.Update(stressStatsMsg{stats: done})
	scr = m.(*stressScreen)
	if scr.started {
		t.Fatal("still started after engine finished")
	}
	if !strings.Contains(scr.View(), "stopped after") {
		t.Fatalf("missing finish summary:\n%s", scr.View())
	}
	if scr.cursor != 4 {
		t.Fatal("cursor should remain on the action row for a restart")
	}
	_, cmd = scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("restart command not issued")
	}
	cmd()
	if !fake.started {
		t.Fatal("engine not restarted")
	}
}

func TestStressInvalidRateShowsError(t *testing.T) {
	scr, fake := newStressScreenForTest(t)
	stressInput(t, scr, 0, "hello")
	stressInput(t, scr, 1, "0")
	scr.cursor = 4
	_, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected no command for invalid rate")
	}
	if !strings.Contains(scr.errMsg, "rate must be a positive number") {
		t.Fatalf("unexpected error %q", scr.errMsg)
	}
	if fake.started {
		t.Fatal("engine started with rate 0")
	}
}

func TestStressDefaultsFromStateMessagesWhenBlank(t *testing.T) {
	scr, fake := newStressScreenForTest(t)
	scr.state.Messages = []string{"from list"}
	stressInput(t, scr, 1, "1")
	scr.cursor = 4
	_, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	cmd()
	if fake.lastConfig.Message != "from list" {
		t.Fatalf("expected first state message, got %q", fake.lastConfig.Message)
	}
}

func TestStressSyncStateWritesBackToAppState(t *testing.T) {
	scr, _ := newStressScreenForTest(t)
	stressInput(t, scr, 0, "abc")
	stressInput(t, scr, 1, "7")
	stressInput(t, scr, 2, "90")
	stressInput(t, scr, 3, "5")
	scr.syncState()
	if scr.state.StressMessage != "abc" || scr.state.StressMessagesPS != 7 ||
		scr.state.StressDuration != 90 || scr.state.StressMaxMsgs != 5 {
		t.Fatalf("state not synced: %+v", scr.state)
	}
}

func TestStressUsesFramingMethodFromState(t *testing.T) {
	scr, fake := newStressScreenForTest(t)
	scr.state.FramingMethod = "non-transparent"
	stressInput(t, scr, 0, "hello")
	stressInput(t, scr, 1, "1")
	scr.cursor = 4
	_, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	cmd()
	if fake.lastConfig.Method != framing.NonTransparent {
		t.Fatalf("method = %q", fake.lastConfig.Method)
	}
}
