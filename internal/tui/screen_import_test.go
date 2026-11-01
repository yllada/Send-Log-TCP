package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newImport(t *testing.T, state *AppState) *importScreen {
	t.Helper()
	s, _ := NewImportScreen(state).(*importScreen)
	_, _ = s.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	s.Init()
	return s
}

func writeImportFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func typePath(s *importScreen, p string) {
	for _, r := range []rune(p) {
		s.path, _ = s.path.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func clearPath(s *importScreen) {
	n := len(s.path.Value())
	for i := 0; i < n; i++ {
		s.path, _ = s.path.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
}

func TestImportAppend(t *testing.T) {
	state := DefaultState()
	state.Messages = []string{"existing"}
	s := newImport(t, state)

	typePath(s, writeImportFile(t, t.TempDir(), "logs.txt", "first line\nsecond line\n"))
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if s.mode != modeReady {
		t.Fatalf("mode: expected ready, got %s", s.mode)
	}
	if len(s.pending) != 2 {
		t.Fatalf("pending: expected 2 messages, got %d", len(s.pending))
	}
	view := s.View()
	if !strings.Contains(view, "imported 2 valid messages") {
		t.Fatalf("preview missing: %s", view)
	}

	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if s.mode != modeDone {
		t.Fatalf("mode after append: expected done, got %s", s.mode)
	}
	if len(state.Messages) != 3 || state.Messages[0] != "existing" || state.Messages[1] != "first line" || state.Messages[2] != "second line" {
		t.Fatalf("append: got %v", state.Messages)
	}
	if !strings.Contains(s.View(), "appended 2") {
		t.Fatalf("done line missing: %s", s.doneText)
	}
}

func TestImportShowsErrors(t *testing.T) {
	state := DefaultState()
	s := newImport(t, state)

	p := writeImportFile(t, t.TempDir(), "logs.json", "not valid json at all\n")
	typePath(s, p)
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if s.mode != modeReady {
		t.Fatalf("mode: expected ready, got %s", s.mode)
	}
	if len(s.preview.Errors) == 0 {
		t.Fatalf("expected import errors to be captured")
	}
	view := s.View()
	if !strings.Contains(view, "invalid JSON") {
		t.Fatalf("error preview missing: %s", view)
	}
}

func TestImportReplace(t *testing.T) {
	state := DefaultState()
	state.Messages = []string{"old1", "old2", "old3"}
	s := newImport(t, state)

	typePath(s, writeImportFile(t, t.TempDir(), "logs.txt", "new one\nnew two\n"))
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if s.mode != modeDone {
		t.Fatalf("mode after replace: expected done, got %s", s.mode)
	}
	if len(state.Messages) != 2 || state.Messages[0] != "new one" || state.Messages[1] != "new two" {
		t.Fatalf("replace: got %v", state.Messages)
	}

	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if s.mode != modeIdle || !s.path.Focused() {
		t.Fatalf("new import: expected idle with focused path, mode=%s", s.mode)
	}
}

func TestImportMissingFile(t *testing.T) {
	state := DefaultState()
	s := newImport(t, state)

	typePath(s, "/no/such/file.txt")
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if s.mode != modeFailed {
		t.Fatalf("mode: expected failed, got %s", s.mode)
	}
	if !s.path.Focused() {
		t.Fatal("path should stay focused to fix the path")
	}
	view := s.View()
	if !strings.Contains(view, "failed to read file") {
		t.Fatalf("error not shown: %s", view)
	}
	if len(state.Messages) != 0 {
		t.Fatalf("state untouched expected, got %v", state.Messages)
	}

	clearPath(s)
	typePath(s, writeImportFile(t, t.TempDir(), "logs.txt", "ok line\n"))
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if s.mode != modeReady {
		t.Fatalf("reimport after fix: expected ready, got %s", s.mode)
	}
}
