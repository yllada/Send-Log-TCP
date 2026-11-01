package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	state    *AppState
	width    int
	height   int
	navIndex int
	current  tea.Model
}

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
	{"Profiles", NewProfilesScreen},
	{"Templates", NewTemplatesScreen},
	{"Config", NewConfigScreen},
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.current != nil {
				m.current = nil
				return m, nil
			}
			return m, tea.Quit
		}
	}

	if m.current != nil {
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
		}
	}
	return m, nil
}

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
	lines = append(lines, "", keyHelp("↑/↓ navigate · enter open · esc quit · ? help"))
	return strings.Join(lines, "\n")
}
