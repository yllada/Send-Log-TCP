// Command sendlog unifies the syslog sender CLI, TUI and GTK GUI in one binary.
package main

import (
	"fmt"
	"os"

	"Sendlog-Syslog/internal/cli"
	"Sendlog-Syslog/internal/gui"
	"Sendlog-Syslog/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		if gui.Available() {
			gui.Run()
			return
		}
		runTUI()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "tui":
		runTUI()
		return
	case "send":
		err = cli.RunSend(args)
	case "connect":
		err = cli.RunConnect(args)
	case "stress":
		err = cli.RunStress(args)
	case "version", "--version", "-v":
		fmt.Println(cli.Version)
	case "help", "--help", "-h":
		fmt.Print(cli.Usage())
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, cli.Usage())
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI() {
	p := tea.NewProgram(tui.New())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
