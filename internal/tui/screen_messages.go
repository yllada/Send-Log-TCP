package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var facilityOptions = []Option{{Value: "0", Label: "0 kern"}, {Value: "1", Label: "1 user"}, {Value: "2", Label: "2 mail"}, {Value: "3", Label: "3 daemon"}, {Value: "4", Label: "4 auth"}, {Value: "5", Label: "5 syslog"}, {Value: "6", Label: "6 lpr"}, {Value: "7", Label: "7 news"}, {Value: "8", Label: "8 uucp"}, {Value: "9", Label: "9 cron"}, {Value: "10", Label: "10 authpriv"}, {Value: "11", Label: "11 ftp"}, {Value: "12", Label: "12 ntp"}, {Value: "13", Label: "13 security"}, {Value: "14", Label: "14 console"}, {Value: "15", Label: "15 sockets"}, {Value: "16", Label: "16 local0"}, {Value: "17", Label: "17 local1"}, {Value: "18", Label: "18 local2"}, {Value: "19", Label: "19 local3"}, {Value: "20", Label: "20 local4"}, {Value: "21", Label: "21 local5"}, {Value: "22", Label: "22 local6"}, {Value: "23", Label: "23 local7"}}

var severityOptions = []Option{{Value: "0", Label: "0 emerg"}, {Value: "1", Label: "1 alert"}, {Value: "2", Label: "2 crit"}, {Value: "3", Label: "3 err"}, {Value: "4", Label: "4 warning"}, {Value: "5", Label: "5 notice"}, {Value: "6", Label: "6 info"}, {Value: "7", Label: "7 debug"}}

type messagesScreen struct {
	state    *AppState
	messages []string
	cursor   int
	listCur  int
	editing  bool
	format   *OptionSelector
	facility *OptionSelector
	severity *OptionSelector
	raw      *OptionSelector
	hostname textinput.Model
	appname  textinput.Model
	editBuf  textinput.Model
}

func NewMessagesScreen(state *AppState) tea.Model {
	hostname := textinput.New()
	appname := textinput.New()
	hostname.Width = 24
	appname.Width = 24
	appname.Placeholder = "sendlog"
	editBuf := textinput.New()
	editBuf.Width = 60
	m := &messagesScreen{
		state:    state,
		messages: append([]string{}, state.Messages...),
		format:   NewOptionSelector([]Option{{Value: "rfc5424", Label: "rfc5424"}, {Value: "rfc3164", Label: "rfc3164"}}),
		facility: NewOptionSelector(facilityOptions),
		severity: NewOptionSelector(severityOptions),
		raw:      NewOptionSelector([]Option{{Value: "off", Label: "off"}, {Value: "on", Label: "on", Helper: "prepend <PRI> only"}}),
		hostname: hostname,
		appname:  appname,
		editBuf:  editBuf,
	}
	m.format.SetByValue(state.Format)
	m.facility.SetByValue(strconv.Itoa(int(state.Facility)))
	m.severity.SetByValue(strconv.Itoa(int(state.Severity)))
	if state.Raw {
		m.raw.SetByValue("on")
	}
	return m
}

func (m *messagesScreen) Init() tea.Cmd {
	return m.hostname.Focus()
}

func (m *messagesScreen) focused() *textinput.Model {
	if m.editing {
		return &m.editBuf
	}
	switch m.cursor {
	case 3:
		return &m.hostname
	case 4:
		return &m.appname
	default:
		return nil
	}
}

func (m *messagesScreen) blur() {
	if m.hostname.Focused() {
		m.hostname.Blur()
	}
	if m.appname.Focused() {
		m.appname.Blur()
	}
	if m.editBuf.Focused() {
		m.editBuf.Blur()
	}
}

func (m *messagesScreen) syncState() {
	if m.editing {
		m.commitEdit()
	}
	m.state.Format = m.format.Value()
	if f, err := strconv.Atoi(m.facility.Value()); err == nil {
		m.state.Facility = uint8(f)
	}
	if sv, err := strconv.Atoi(m.severity.Value()); err == nil {
		m.state.Severity = uint8(sv)
	}
	m.state.Hostname = m.hostname.Value()
	m.state.Appname = m.appname.Value()
	m.state.Raw = m.raw.Value() == "on"
	m.state.Messages = m.messages
}

func (m *messagesScreen) stepSel(sel *OptionSelector, k rune) {
	switch k {
	case 'l':
		sel.Toggle()
	case 'h':
		sel.Up()
	case 'j':
		sel.Down()
	}
}

func (m *messagesScreen) startEdit() {
	m.editing = true
	m.editBuf.SetValue(m.messages[m.listCur])
	m.editBuf.CursorEnd()
}

func (m *messagesScreen) commitEdit() {
	m.messages[m.listCur] = m.editBuf.Value()
	m.editing = false
}

func (m *messagesScreen) moveCursor(delta int) tea.Cmd {
	if m.editing {
		m.commitEdit()
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 6
	}
	if m.cursor > 6 {
		m.cursor = 0
	}
	m.blur()
	if in := m.focused(); in != nil {
		return in.Focus()
	}
	return nil
}

func (m *messagesScreen) moveList(delta int) {
	if m.editing {
		m.commitEdit()
	}
	if len(m.messages) == 0 {
		return
	}
	m.listCur += delta
	if m.listCur < 0 {
		m.listCur = len(m.messages) - 1
	}
	if m.listCur >= len(m.messages) {
		m.listCur = 0
	}
}

func (m *messagesScreen) addMessage() tea.Cmd {
	m.messages = append(m.messages, "")
	m.listCur = len(m.messages) - 1
	m.startEdit()
	return m.editBuf.Focus()
}

func (m *messagesScreen) deleteMessage() {
	if m.editing {
		m.commitEdit()
	}
	if len(m.messages) == 0 {
		return
	}
	m.messages = append(m.messages[:m.listCur], m.messages[m.listCur+1:]...)
	if m.listCur >= len(m.messages) {
		m.listCur = len(m.messages) - 1
	}
	if m.listCur < 0 {
		m.listCur = 0
	}
}

func (m *messagesScreen) listKey(k rune) tea.Cmd {
	switch k {
	case 'k':
		m.moveList(-1)
	case 'j':
		m.moveList(1)
	case 'a':
		return m.addMessage()
	case 'x':
		m.deleteMessage()
	}
	return nil
}

func (m *messagesScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyMsg:
		if in := m.focused(); in != nil {
			switch msg.Type {
			case tea.KeyEnter:
				if m.editing {
					m.commitEdit()
					return m, nil
				}
				return m, m.moveCursor(1)
			case tea.KeyTab:
				return m, m.moveCursor(1)
			case tea.KeyShiftTab:
				return m, m.moveCursor(-1)
			case tea.KeyRunes, tea.KeyBackspace, tea.KeySpace:
				var cmd tea.Cmd
				*in, cmd = in.Update(msg)
				return m, cmd
			}
			return m, nil
		}
		switch msg.Type {
		case tea.KeyEnter:
			if m.cursor == 6 {
				if len(m.messages) > 0 {
					m.startEdit()
					return m, m.editBuf.Focus()
				}
				return m, m.addMessage()
			}
			return m, m.moveCursor(1)
		case tea.KeyTab:
			return m, m.moveCursor(1)
		case tea.KeyShiftTab:
			return m, m.moveCursor(-1)
		case tea.KeyRunes:
			if len(msg.Runes) != 1 {
				return m, nil
			}
			k := msg.Runes[0]
			switch m.cursor {
			case 0:
				m.stepSel(m.format, k)
			case 1:
				m.stepSel(m.facility, k)
			case 2:
				m.stepSel(m.severity, k)
			case 5:
				m.stepSel(m.raw, k)
			case 6:
				return m, m.listKey(k)
			}
		}
	}
	return m, nil
}

func (m *messagesScreen) View() string {
	if in := m.focused(); in != nil {
		in.Cursor.Blink = false
	}
	var b strings.Builder
	b.WriteString(titleStyle("Messages"))
	b.WriteString("\n\n")
	b.WriteString(labelStyle("Format:   ") + m.format.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Facility: ") + m.facility.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Severity: ") + m.severity.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Hostname: ") + m.hostname.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Appname:  ") + m.appname.View())
	b.WriteString("\n")
	b.WriteString(labelStyle("Raw PRI:  ") + m.raw.View())
	b.WriteString("\n\n")
	if m.editing {
		b.WriteString(labelStyle("edit: ") + m.editBuf.View())
	} else if len(m.messages) == 0 {
		b.WriteString(dim.Render("  (no messages — enter to add)"))
	} else {
		for i, msg := range m.messages {
			marker := "  "
			if i == m.listCur {
				marker = "\u25b8 "
			}
			fmt.Fprintf(&b, "%s%2d. %s\n", marker, i+1, msg)
		}
	}
	b.WriteString("\n")
	b.WriteString(keyHelp("tab/enter next · backtab prev · l toggle · h/j step · enter edit · a add · x delete · esc back"))
	return screenBox("Messages", b.String())
}
