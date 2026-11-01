package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	accent = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	ok     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	err    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func titleStyle(text string) string { return accent.Bold(true).Render(text) }
func labelStyle(text string) string { return dim.Render(text) }
func keyHelp(text string) string    { return dim.Render(text) }
func okText(text string) string     { return ok.Render(text) }
func errText(text string) string    { return err.Render(text) }

func screenBox(title, content string) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
	return box.Render(lipgloss.JoinVertical(lipgloss.Left, titleStyle(title), content))
}
