package tui

import tea "github.com/charmbracelet/bubbletea"

type templatesScreen struct{ state *AppState }

func NewTemplatesScreen(state *AppState) tea.Model { return &templatesScreen{state: state} }
func (t *templatesScreen) Init() tea.Cmd           { return nil }
func (t *templatesScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return t, nil
}
func (t *templatesScreen) View() string {
	return screenBox("Templates", "TODO T5: list, save, apply, delete")
}
