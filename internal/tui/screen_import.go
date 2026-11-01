package tui

import (
	"fmt"
	"strings"

	"Sendlog-Syslog/internal/logimport"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type importMode string

const (
	modeIdle   importMode = "idle"
	modeReady  importMode = "ready"
	modeDone   importMode = "done"
	modeFailed importMode = "failed"
)

type importScreen struct {
	state    *AppState
	path     textinput.Model
	mode     importMode
	preview  logimport.Result
	pending  []string
	errMsg   string
	doneText string
}

func NewImportScreen(state *AppState) tea.Model {
	path := textinput.New()
	path.Width = 50
	path.Placeholder = "/path/to/logs.txt"
	return &importScreen{state: state, path: path, mode: modeIdle}
}

func (i *importScreen) Init() tea.Cmd {
	return i.path.Focus()
}

func (i *importScreen) syncState() {}

func (i *importScreen) runImport() {
	result, err := logimport.Import(i.path.Value())
	if err != nil {
		i.mode = modeFailed
		i.errMsg = err.Error()
		i.pending = nil
		return
	}
	i.preview = result
	i.pending = result.Messages
	i.mode = modeReady
	i.errMsg = ""
	i.path.Blur()
}

func (i *importScreen) apply(decision string) {
	switch decision {
	case "append":
		i.state.Messages = append(i.state.Messages, i.pending...)
		i.doneText = fmt.Sprintf("appended %d messages (%d total)", len(i.pending), len(i.state.Messages))
	case "replace":
		i.state.Messages = append([]string{}, i.pending...)
		i.doneText = fmt.Sprintf("replaced all: now %d messages", len(i.state.Messages))
	}
	i.mode = modeDone
}

func (i *importScreen) resetForNewImport() tea.Cmd {
	i.path.SetValue("")
	i.mode = modeIdle
	return i.path.Focus()
}

func (i *importScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return i, nil
	case tea.KeyMsg:
		if i.path.Focused() {
			switch msg.Type {
			case tea.KeyEnter:
				i.runImport()
				return i, nil
			case tea.KeyTab:
				i.path.Blur()
				return i, nil
			case tea.KeyRunes, tea.KeyBackspace, tea.KeySpace:
				var cmd tea.Cmd
				i.path, cmd = i.path.Update(msg)
				return i, cmd
			}
			return i, nil
		}
		if msg.Type == tea.KeyEnter {
			if i.mode == modeDone {
				return i, i.resetForNewImport()
			}
			if i.mode == modeFailed {
				return i, i.path.Focus()
			}
			return i, nil
		}
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			switch i.mode {
			case modeReady:
				switch msg.Runes[0] {
				case 'a':
					i.apply("append")
				case 'r':
					i.apply("replace")
				case 'f':
					return i, i.path.Focus()
				}
			case modeDone:
				if msg.Runes[0] == 'i' {
					return i, i.resetForNewImport()
				}
			case modeFailed, modeIdle:
				if msg.Runes[0] == 'f' {
					return i, i.path.Focus()
				}
			}
		}
	}
	return i, nil
}

func (i *importScreen) View() string {
	if i.path.Focused() {
		i.path.Cursor.Blink = false
	}
	var b strings.Builder
	b.WriteString(titleStyle("Import File"))
	b.WriteString("\n\n")
	b.WriteString(labelStyle("Path: ") + i.path.View())
	b.WriteString("\n\n")
	switch i.mode {
	case modeReady:
		b.WriteString(okText(fmt.Sprintf("imported %d valid messages from %d lines", len(i.preview.Messages), i.preview.TotalLines)))
		b.WriteString("\n")
		for j, e := range i.preview.Errors {
			if j >= 5 {
				break
			}
			if len(e) > 60 {
				e = e[:57] + "..."
			}
			b.WriteString(errText("  " + e))
			b.WriteString("\n")
		}
		if len(i.preview.Errors) > 5 {
			fmt.Fprintf(&b, "  +%d more errors\n", len(i.preview.Errors)-5)
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "%s %d messages\n", labelStyle("current list:"), len(i.state.Messages))
		b.WriteString(keyHelp("a append · r replace all · f edit path"))
	case modeFailed:
		b.WriteString(errText(i.errMsg))
	case modeDone:
		b.WriteString(okText(i.doneText))
	case modeIdle:
		b.WriteString(dim.Render("type a path and press enter to import"))
	}
	b.WriteString("\n\n")
	b.WriteString(keyHelp("enter import · a append · r replace all · i new import · esc back"))
	return screenBox("Import File", b.String())
}
