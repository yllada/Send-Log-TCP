package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type connectionScreen struct{ state *AppState }

func NewConnectionScreen(state *AppState) tea.Model {
	return &connectionScreen{state: state}
}

func (c *connectionScreen) Init() tea.Cmd { return nil }

func (c *connectionScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

func (c *connectionScreen) View() string {
	return screenBox("Connection", "TODO T2: connection form + test connection")
}
