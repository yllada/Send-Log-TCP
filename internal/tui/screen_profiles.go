package tui

import tea "github.com/charmbracelet/bubbletea"

type profilesScreen struct{ state *AppState }

func NewProfilesScreen(state *AppState) tea.Model { return &profilesScreen{state: state} }
func (p *profilesScreen) Init() tea.Cmd          { return nil }
func (p *profilesScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return p, nil
}
func (p *profilesScreen) View() string {
	return screenBox("Profiles", "TODO T5: list, save, load, delete")
}
