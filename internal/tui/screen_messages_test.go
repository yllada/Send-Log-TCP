package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newMessages(t *testing.T, state *AppState) *messagesScreen {
	t.Helper()
	m, _ := NewMessagesScreen(state).(*messagesScreen)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m.Init()
	return m
}

func mkey(runes ...rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
}

func tab() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyTab}
}

func enter() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyEnter}
}

func TestMessagesAddDelete(t *testing.T) {
	state := DefaultState()
	m := newMessages(t, state)

	for i := 0; i < 6; i++ {
		_, _ = m.Update(tab())
	}
	if m.cursor != 6 {
		t.Fatalf("cursor: expected 6, got %d", m.cursor)
	}

	_, _ = m.Update(enter())
	for _, r := range "hero" {
		_, _ = m.Update(mkey(r))
	}
	_, _ = m.Update(enter())
	if got := m.messages; len(got) != 1 || got[0] != "hero" {
		t.Fatalf("add: expected [hero], got %v", got)
	}

	_, _ = m.Update(mkey('x'))
	if len(m.messages) != 0 {
		t.Fatalf("delete: expected empty list, got %v", m.messages)
	}
	if m.listCur != 0 {
		t.Fatalf("listCur after deleting last: expected 0, got %d", m.listCur)
	}
	view := m.View()
	if !strings.Contains(view, "no messages") {
		t.Fatalf("view should show empty hint, got:\n%s", view)
	}
}

func TestMessagesEdit(t *testing.T) {
	state := DefaultState()
	state.Messages = []string{"one", "two"}
	m := newMessages(t, state)
	m.listCur = 1
	for i := 0; i < 6; i++ {
		_, _ = m.Update(tab())
	}
	_, _ = m.Update(enter())
	if !m.editing {
		t.Fatal("enter on a list item should start editing")
	}
	for i := 0; i < 3; i++ {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}
	for _, r := range "edit" {
		_, _ = m.Update(mkey(r))
	}
	if got := m.editBuf.Value(); got != "edit" {
		t.Fatalf("editBuf: expected edit, got %q", got)
	}
	_, _ = m.Update(enter())
	if m.editing {
		t.Fatal("enter should commit the edit")
	}
	if got := m.messages; got[0] != "one" || got[1] != "edit" {
		t.Fatalf("edit: expected [one edit], got %v", got)
	}
}

func TestMessagesToggleRaw(t *testing.T) {
	state := DefaultState()
	m := newMessages(t, state)
	if m.raw.Value() != "off" {
		t.Fatalf("raw default: expected off, got %s", m.raw.Value())
	}
	for i := 0; i < 5; i++ {
		_, _ = m.Update(tab())
	}
	if m.cursor != 5 {
		t.Fatalf("cursor: expected 5, got %d", m.cursor)
	}
	_, _ = m.Update(mkey('l'))
	if m.raw.Value() != "on" {
		t.Fatalf("raw toggle: expected on, got %s", m.raw.Value())
	}
}

func TestMessagesFacilitySeverity(t *testing.T) {
	state := DefaultState()
	m := newMessages(t, state)
	if m.facility.Value() != "13" {
		t.Fatalf("facility default: expected 13, got %s", m.facility.Value())
	}
	_, _ = m.Update(tab())
	if m.cursor != 1 {
		t.Fatalf("cursor: expected 1, got %d", m.cursor)
	}
	_, _ = m.Update(mkey('h'))
	_, _ = m.Update(mkey('h'))
	if m.facility.Value() != "11" {
		t.Fatalf("facility step: expected 11, got %s", m.facility.Value())
	}
	_, _ = m.Update(tab())
	if m.cursor != 2 {
		t.Fatalf("cursor: expected 2, got %d", m.cursor)
	}
	_, _ = m.Update(mkey('j'))
	if m.severity.Value() != "7" {
		t.Fatalf("severity step: expected 7, got %s", m.severity.Value())
	}
}

func TestMessagesSyncState(t *testing.T) {
	state := DefaultState()
	m := newMessages(t, state)

	_, _ = m.Update(mkey('l'))
	_, _ = m.Update(tab())
	_, _ = m.Update(mkey('j'))
	_, _ = m.Update(tab())
	_, _ = m.Update(mkey('h'))
	_, _ = m.Update(tab())
	for _, r := range "host" {
		_, _ = m.Update(mkey(r))
	}
	_, _ = m.Update(tab())
	for _, r := range "app" {
		_, _ = m.Update(mkey(r))
	}
	_, _ = m.Update(tab())
	_, _ = m.Update(mkey('l'))
	_, _ = m.Update(tab())
	_, _ = m.Update(mkey('a'))
	for _, r := range "msg" {
		_, _ = m.Update(mkey(r))
	}
	_, _ = m.Update(enter())

	m.syncState()
	if state.Format != "rfc3164" {
		t.Errorf("Format: expected rfc3164, got %s", state.Format)
	}
	if state.Facility != 14 {
		t.Errorf("Facility: expected 14, got %d", state.Facility)
	}
	if state.Severity != 5 {
		t.Errorf("Severity: expected 5, got %d", state.Severity)
	}
	if state.Hostname != "host" {
		t.Errorf("Hostname: expected host, got %s", state.Hostname)
	}
	if state.Appname != "app" {
		t.Errorf("Appname: expected app, got %s", state.Appname)
	}
	if !state.Raw {
		t.Error("Raw: expected true")
	}
	if len(state.Messages) != 1 || state.Messages[0] != "msg" {
		t.Errorf("Messages: expected [msg], got %v", state.Messages)
	}
}

func TestMessagesInitialState(t *testing.T) {
	state := DefaultState()
	state.Format = "rfc3164"
	state.Facility = 1
	state.Severity = 3
	state.Raw = true
	m := newMessages(t, state)
	if m.format.Value() != "rfc3164" {
		t.Errorf("format: expected rfc3164, got %s", m.format.Value())
	}
	if m.facility.Value() != "1" {
		t.Errorf("facility: expected 1, got %s", m.facility.Value())
	}
	if m.severity.Value() != "3" {
		t.Errorf("severity: expected 3, got %s", m.severity.Value())
	}
	if m.raw.Value() != "on" {
		t.Errorf("raw: expected on, got %s", m.raw.Value())
	}
}
