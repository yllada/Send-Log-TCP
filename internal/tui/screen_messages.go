package tui

import tea "github.com/charmbracelet/bubbletea"

type messagesScreen struct{ state *AppState }

func NewMessagesScreen(state *AppState) tea.Model { return &messagesScreen{state: state} }
func (m *messagesScreen) Init() tea.Cmd           { return nil }
func (m *messagesScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
func (m *messagesScreen) View() string {
	return screenBox("Messages", "TODO T3: format, facility, severity, messages list")
}
