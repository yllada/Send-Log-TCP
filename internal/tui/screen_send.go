package tui

import tea "github.com/charmbracelet/bubbletea"

type sendScreen struct{ state *AppState }

func NewSendScreen(state *AppState) tea.Model { return &sendScreen{state: state} }
func (s *sendScreen) Init() tea.Cmd          { return nil }
func (s *sendScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}
func (s *sendScreen) View() string {
	return screenBox("Send", "TODO T3: send current messages, per-message results")
}
