package tui

import tea "github.com/charmbracelet/bubbletea"

type stressScreen struct{ state *AppState }

func NewStressScreen(state *AppState) tea.Model { return &stressScreen{state: state} }
func (s *stressScreen) Init() tea.Cmd           { return nil }
func (s *stressScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}
func (s *stressScreen) View() string {
	return screenBox("Stress", "TODO T4: rate, duration, start/stop, live stats")
}
