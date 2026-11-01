package tui

import tea "github.com/charmbracelet/bubbletea"

type importScreen struct{ state *AppState }

func NewImportScreen(state *AppState) tea.Model { return &importScreen{state: state} }
func (i *importScreen) Init() tea.Cmd           { return nil }
func (i *importScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return i, nil
}
func (i *importScreen) View() string {
	return screenBox("Import File", "TODO T3: file path, result preview, append/replace")
}
