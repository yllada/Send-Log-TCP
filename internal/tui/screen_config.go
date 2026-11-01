package tui

import tea "github.com/charmbracelet/bubbletea"

type configScreen struct{ state *AppState }

func NewConfigScreen(state *AppState) tea.Model { return &configScreen{state: state} }
func (c *configScreen) Init() tea.Cmd           { return nil }
func (c *configScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}
func (c *configScreen) View() string {
	return screenBox("Config", "TODO T6: export/import profiles+templates")
}
