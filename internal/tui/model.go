// Package tui implements the Bubble Tea terminal front end over the shared syslog domain state.
package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// stateSyncer is implemented by screens that write their working fields back to
// the shared state; the root calls it before closing so nothing is lost.
type stateSyncer interface{ syncState() }

// Model is the root Bubble Tea model that owns the shared app state and delegates to the active screen.
type Model struct {
	state    *AppState
	width    int
	height   int
	navIndex int
	current  tea.Model
}

// New builds the root model seeded with default state.
func New() *Model {
	return &Model{state: DefaultState()}
}

var screens = []struct {
	label string
	new   func(*AppState) tea.Model
}{
	{"Connection", NewConnectionScreen},
	{"Messages", NewMessagesScreen},
	{"Import File", NewImportScreen},
	{"Send", NewSendScreen},
	{"Stress", NewStressScreen},
}

// Init implements tea.Model; the root model needs no startup command.
func (m *Model) Init() tea.Cmd { return nil }

// Update handles root-level keys (size, quit, back) and forwards the rest to the active screen.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.current != nil {
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
			if syncer, ok := m.current.(stateSyncer); ok {
				syncer.syncState()
			}
			m.current = nil
			return m, nil
		}
		var cmd tea.Cmd
		m.current, cmd = m.current.Update(msg)
		return m, cmd
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			m.navIndex = (m.navIndex + len(screens) - 1) % len(screens)
		case "down", "j":
			m.navIndex = (m.navIndex + 1) % len(screens)
		case "enter":
			m.current = screens[m.navIndex].new(m.state)
			return m, m.current.Init()
		case "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the active screen, or the screen picker when none is open.
func (m *Model) View() string {
	if m.current != nil {
		return m.current.View()
	}
	return m.viewNav()
}

func (m *Model) viewNav() string {
	lines := []string{"", titleStyle("SendLog")}
	for i, s := range screens {
		if i == m.navIndex {
			lines = append(lines, accent.Render("> ")+s.label)
		} else {
			lines = append(lines, dim.Render("  "+s.label))
		}
	}
	lines = append(lines, "", keyHelp("↑/↓ navigate · enter open · esc quit"))
	return screenBox("SendLog", strings.Join(lines, "\n"))
}
