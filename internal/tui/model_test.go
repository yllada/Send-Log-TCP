package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func update(t *testing.T, m *Model, msg tea.Msg) *Model {
	t.Helper()
	next, _ := m.Update(msg)
	m, ok := next.(*Model)
	if !ok {
		t.Fatal("Update returned wrong model type")
	}
	return m
}

func TestNewStartsInNav(t *testing.T) {
	m := New()
	if m.current != nil {
		t.Fatal("expected to start in navigation")
	}
	if m.navIndex != 0 {
		t.Fatalf("expected nav index 0, got %d", m.navIndex)
	}
}

func TestNavMovesWrapAround(t *testing.T) {
	m := New()

	m = update(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.navIndex != 1 {
		t.Fatalf("down: expected 1, got %d", m.navIndex)
	}

	for i := 0; i < len(screens)-1; i++ {
		m = update(t, m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.navIndex != 0 {
		t.Fatalf("wrap: expected 0, got %d", m.navIndex)
	}

	m = update(t, m, tea.KeyMsg{Type: tea.KeyUp})
	if m.navIndex != len(screens)-1 {
		t.Fatalf("wrap up: expected %d, got %d", len(screens)-1, m.navIndex)
	}
}

func TestEnterOpensSelectedScreen(t *testing.T) {
	m := New()
	m = update(t, m, tea.KeyMsg{Type: tea.KeyDown})
	m = update(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.current == nil {
		t.Fatal("enter should open a screen")
	}
}

func TestEscClosesScreenAndQuitsFromNav(t *testing.T) {
	m := New()
	m = update(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m = update(t, m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.current != nil {
		t.Fatal("esc should close the screen")
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc in nav should return the quit command")
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should return the quit command")
	}
}
