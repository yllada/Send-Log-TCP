package main

import (
	"log"

	"Sendlog-Syslog/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(tui.New())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
