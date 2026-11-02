//go:build linux && cgo

package gui

import (
	"context"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var facilityNames = []string{
	"kern", "user", "mail", "daemon", "auth", "syslog", "lpr", "news",
	"uucp", "cron", "authpriv", "ftp", "ntp", "security", "console", "sockets",
	"local0", "local1", "local2", "local3", "local4", "local5", "local6", "local7",
}

var severityNames = []string{
	"emergency", "alert", "critical", "error",
	"warning", "notice", "informational", "debug",
}

// field wraps a widget in a vertical box with its label on top,
// matching the Kiwi "label above input" pattern.
func field(label string, child gtk.Widgetter) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 4)
	l := gtk.NewLabel(label)
	l.SetXAlign(0.0)
	l.AddCSSClass("sendlog-field-label")
	box.Append(l)
	box.Append(child)
	return box
}

// switchField wraps a switch in the label-on-top pattern, keeping the
// switch left-aligned within its cell so it never stretches.
func switchField(label string, sw *gtk.Switch) *gtk.Box {
	sw.SetHAlign(gtk.AlignStart)
	box := gtk.NewBox(gtk.OrientationVertical, 4)
	l := gtk.NewLabel(label)
	l.SetXAlign(0.0)
	l.AddCSSClass("sendlog-field-label")
	box.Append(l)
	box.Append(sw)
	return box
}

func (a *App) sectionFrame(title, subtitle string) (*gtk.Frame, *gtk.Grid) {
	frame := gtk.NewFrame(title)
	frame.AddCSSClass("sendlog-section")

	grid := gtk.NewGrid()
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(10)
	grid.SetColumnHomogeneous(true)
	grid.SetMarginTop(12)
	grid.SetMarginBottom(12)
	grid.SetMarginStart(12)
	grid.SetMarginEnd(12)
	if subtitle != "" {
		sub := gtk.NewLabel(subtitle)
		sub.SetXAlign(0.0)
		sub.AddCSSClass("sendlog-section-sub")
		grid.Attach(sub, 0, 0, 3, 1)
	}
	return frame, grid
}

// pickPathField pairs an entry with an embedded folder icon that opens
// the native file chooser and writes the path in.
// It returns the field box (for sensitivity) and the entry.
func (a *App) pickPathField(label, placeholder string, set func(string)) (*gtk.Box, *gtk.Entry) {
	entry := gtk.NewEntry()
	entry.SetPlaceholderText(placeholder)
	entry.ConnectChanged(func() { set(entry.Text()) })

	entry.SetIconFromIconName(gtk.EntryIconSecondary, "folder-open-symbolic")
	entry.SetIconActivatable(gtk.EntryIconSecondary, true)
	entry.SetIconTooltipText(gtk.EntryIconSecondary, "Choose file…")
	entry.ConnectIconPress(func(iconPos gtk.EntryIconPosition) {
		if iconPos != gtk.EntryIconSecondary {
			return
		}
		dialog := gtk.NewFileDialog()
		dialog.SetTitle(label)
		dialog.SetModal(true)
		dialog.Open(context.Background(), &a.win.Window, func(asyncRes gio.AsyncResulter) {
			file, err := dialog.OpenFinish(asyncRes)
			if err != nil || file == nil {
				return
			}
			path := file.Path()
			glib.IdleAdd(func() { entry.SetText(path) })
		})
	})

	box := gtk.NewBox(gtk.OrientationVertical, 4)
	l := gtk.NewLabel(label)
	l.SetXAlign(0.0)
	l.AddCSSClass("sendlog-field-label")
	box.Append(l)
	box.Append(entry)
	return box, entry
}

func (a *App) buildConnectionFrame() *gtk.Frame {
	frame, grid := a.sectionFrame("Connection", "Target syslog server and transport")
	row := 1

	a.connEntry = gtk.NewEntry()
	a.connEntry.SetPlaceholderText("hostname or IP")
	a.connEntry.SetText(a.state.Address)
	a.connEntry.ConnectChanged(func() { a.state.Address = a.connEntry.Text() })

	a.portEntry = gtk.NewEntry()
	a.portEntry.SetText(a.state.Port)
	a.portEntry.SetWidthChars(6)
	a.portEntry.ConnectChanged(func() { a.state.Port = a.portEntry.Text() })

	a.protoCombo = gtk.NewComboBoxText()
	a.protoCombo.Append("tcp", "TCP")
	a.protoCombo.Append("udp", "UDP")
	a.protoCombo.SetActive(0)
	a.protoCombo.ConnectChanged(func() {
		a.state.Protocol = a.protoCombo.ActiveID()
		a.updateFramingRows()
	})

	grid.Attach(field("Address", a.connEntry), 0, row, 1, 1)
	grid.Attach(field("Port", a.portEntry), 1, row, 1, 1)
	grid.Attach(field("Protocol", a.protoCombo), 2, row, 1, 1)
	row++

	a.frameCombo = gtk.NewComboBoxText()
	a.frameCombo.Append("octet-counting", "Octet-counting (RFC 6587)")
	a.frameCombo.Append("non-transparent", "Non-transparent")
	a.frameCombo.SetActive(0)
	a.frameCombo.ConnectChanged(func() { a.state.FramingMethod = a.frameCombo.ActiveID() })

	a.tlsSwitch = gtk.NewSwitch()
	a.tlsSwitch.SetActive(a.state.UseTLS)
	a.tlsSwitch.ConnectStateSet(func(state bool) bool {
		a.state.UseTLS = state
		a.updateTLSRows()
		return false
	})

	a.verifySw = gtk.NewSwitch()
	a.verifySw.SetActive(a.state.TLSVerify)
	a.verifySw.ConnectStateSet(func(state bool) bool {
		a.state.TLSVerify = state
		return false
	})

	grid.Attach(field("Framing", a.frameCombo), 0, row, 1, 1)
	grid.Attach(switchField("TLS", a.tlsSwitch), 1, row, 1, 1)
	grid.Attach(switchField("Verify certificates", a.verifySw), 2, row, 1, 1)
	row++

	caField, entryCA := a.pickPathField("CA certificate", "CA certificate path (optional)", func(p string) { a.state.CACertPath = p })
	certField, entryCert := a.pickPathField("Client certificate", "client certificate path (optional)", func(p string) { a.state.ClientCertPath = p })
	keyField, entryKey := a.pickPathField("Client key", "client key path (optional)", func(p string) { a.state.ClientKeyPath = p })
	a.caEntry, a.certEntry, a.keyEntry = entryCA, entryCert, entryKey
	grid.Attach(caField, 0, row, 1, 1)
	grid.Attach(certField, 1, row, 1, 1)
	grid.Attach(keyField, 2, row, 1, 1)
	row++

	a.connToggle = gtk.NewToggleButtonWithLabel("Connect")
	a.connToggle.SetTooltipText("Open or close the syslog connection")
	a.connToggle.SetHExpand(true)
	a.connToggle.ConnectToggled(func() { a.onToggleConnect() })
	grid.Attach(a.connToggle, 0, row, 3, 1)
	row++

	a.updateTLSRows()
	a.updateFramingRows()
	frame.SetChild(grid)
	return frame
}

func (a *App) updateTLSRows() {
	a.verifySw.SetSensitive(a.state.UseTLS)
	a.caEntry.SetSensitive(a.state.UseTLS)
	a.certEntry.SetSensitive(a.state.UseTLS)
	a.keyEntry.SetSensitive(a.state.UseTLS)
}

// updateFramingRows disables the framing picker for UDP, where datagrams
// delimit messages and framing does not apply.
func (a *App) updateFramingRows() {
	a.frameCombo.SetSensitive(a.state.Protocol == "tcp")
}

func (a *App) buildMessageFrame() *gtk.Frame {
	frame, grid := a.sectionFrame("Message", "Syslog header fields and payload")
	row := 1

	a.formatCombo = gtk.NewComboBoxText()
	a.formatCombo.Append("rfc5424", "RFC 5424")
	a.formatCombo.Append("rfc3164", "RFC 3164")
	a.formatCombo.Append("raw-pri", "Raw PRI (replay)")
	a.formatCombo.SetActive(0)
	a.formatCombo.ConnectChanged(func() {
		a.state.Format = a.formatCombo.ActiveID()
		a.state.Raw = a.state.Format == "raw-pri"
	})

	a.facCombo = gtk.NewComboBoxText()
	for i, name := range facilityNames {
		a.facCombo.Append(strconv.Itoa(i), strconv.Itoa(i)+" "+name)
	}
	a.facCombo.SetActive(int(a.state.Facility))
	a.facCombo.ConnectChanged(func() { a.state.Facility = uint8(a.facCombo.Active()) })

	a.sevCombo = gtk.NewComboBoxText()
	for i, name := range severityNames {
		a.sevCombo.Append(strconv.Itoa(i), strconv.Itoa(i)+" "+name)
	}
	a.sevCombo.SetActive(int(a.state.Severity))
	a.sevCombo.ConnectChanged(func() { a.state.Severity = uint8(a.sevCombo.Active()) })

	a.msgBuf = gtk.NewTextBuffer(nil)
	a.msgView = gtk.NewTextViewWithBuffer(a.msgBuf)
	a.msgView.SetEditable(true)
	msgScroll := gtk.NewScrolledWindow()
	msgScroll.AddCSSClass("sendlog-msg-view")
	msgScroll.SetChild(a.msgView)
	msgScroll.SetSizeRequest(-1, 80)
	a.msgBuf.ConnectChanged(func() { a.state.Messages = a.readMessages() })

	grid.Attach(field("Message(s) — one per line", msgScroll), 0, row, 3, 1)
	row++

	importRow := gtk.NewBox(gtk.OrientationHorizontal, 8)
	importBtn := gtk.NewButtonWithLabel("Import file…")
	importBtn.ConnectClicked(func() { a.importFile() })
	importRow.Append(importBtn)
	a.importLabel = gtk.NewLabel("No file imported")
	a.importLabel.SetHAlign(gtk.AlignStart)
	a.importLabel.AddCSSClass("sendlog-imported-badge")
	importRow.Append(a.importLabel)
	grid.Attach(field("Batch import", importRow), 0, row, 3, 1)
	row++

	a.sendBtn = gtk.NewButtonWithLabel("Send messages")
	a.sendBtn.SetHExpand(true)
	a.sendBtn.ConnectClicked(func() { a.onSend() })
	a.sendBtn.SetSensitive(false)
	grid.Attach(a.sendBtn, 0, row, 3, 1)
	row++

	a.applyDefaultsToWidgets()
	frame.SetChild(grid)
	return frame
}

func (a *App) buildStressFrame() *gtk.Frame {
	frame, grid := a.sectionFrame("Stress", "Continuous sending for load testing")
	row := 1

	a.stressMsgEnt = gtk.NewEntry()
	a.stressMsgEnt.SetText(a.state.StressMessage)
	a.stressMsgEnt.ConnectChanged(func() { a.state.StressMessage = a.stressMsgEnt.Text() })
	grid.Attach(field("Message", a.stressMsgEnt), 0, row, 3, 1)
	row++

	a.rateSpin = gtk.NewSpinButtonWithRange(1, 100000, 1)
	a.rateSpin.SetValue(float64(a.state.StressRate))
	a.rateSpin.ConnectValueChanged(func() { a.state.StressRate = a.rateSpin.ValueAsInt() })

	a.durSpin = gtk.NewSpinButtonWithRange(0, 86400, 1)
	a.durSpin.SetValue(float64(a.state.StressDuration))
	a.durSpin.ConnectValueChanged(func() { a.state.StressDuration = a.durSpin.ValueAsInt() })

	a.maxSpin = gtk.NewSpinButtonWithRange(0, 1000000, 1)
	a.maxSpin.SetValue(float64(a.state.StressMax))
	a.maxSpin.ConnectValueChanged(func() { a.state.StressMax = a.maxSpin.ValueAsInt() })

	grid.Attach(field("Rate (msg/s)", a.rateSpin), 0, row, 1, 1)
	grid.Attach(field("Duration (s, 0 = until stop)", a.durSpin), 1, row, 1, 1)
	grid.Attach(field("Max messages (0 = unlimited)", a.maxSpin), 2, row, 1, 1)
	row++

	a.randSwitch = gtk.NewSwitch()
	a.randSwitch.SetActive(a.state.RandomizeData)
	a.randSwitch.ConnectStateSet(func(state bool) bool {
		a.state.RandomizeData = state
		return false
	})
	grid.Attach(switchField("Randomize data", a.randSwitch), 0, row, 1, 1)
	row++

	btnRow := gtk.NewBox(gtk.OrientationHorizontal, 8)
	a.stressBtn = gtk.NewButtonWithLabel("Start stress")
	a.stressBtn.ConnectClicked(func() { a.onStartStress() })
	a.stressBtn.SetSensitive(false)
	btnRow.Append(a.stressBtn)
	a.stopBtn = gtk.NewButtonWithLabel("Stop")
	a.stopBtn.ConnectClicked(func() { a.onStop() })
	a.stopBtn.SetSensitive(false)
	btnRow.Append(a.stopBtn)
	grid.Attach(btnRow, 0, row, 3, 1)
	row++

	frame.SetChild(grid)
	return frame
}

func (a *App) buildStatusArea() *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 8)
	box.AddCSSClass("sendlog-status")

	statusLine := gtk.NewBox(gtk.OrientationHorizontal, 8)
	a.statusIcon = gtk.NewLabel("•")
	a.statusIcon.AddCSSClass("sendlog-status-icon")
	a.statusText = gtk.NewLabel("Not connected")
	a.statusText.SetXAlign(0.0)
	a.statusText.AddCSSClass("sendlog-status-text")
	statusLine.Append(a.statusIcon)
	statusLine.Append(a.statusText)
	box.Append(statusLine)

	a.statusBar = gtk.NewProgressBar()
	a.statusBar.SetFraction(0)
	a.statusBar.SetShowText(false)
	box.Append(a.statusBar)

	a.details = gtk.NewTextView()
	a.details.SetEditable(false)
	a.details.AddCSSClass("sendlog-monospace")
	detailsScroll := gtk.NewScrolledWindow()
	detailsScroll.SetChild(a.details)
	detailsScroll.SetSizeRequest(-1, 100)
	box.Append(detailsScroll)

	note := gtk.NewLabel("Stress stats appear here while a run is in progress")
	note.SetXAlign(0.0)
	note.AddCSSClass("sendlog-footer-note")
	box.Append(note)
	return box
}

// applyDefaultsToWidgets seeds the message frame with the current defaults.
func (a *App) applyDefaultsToWidgets() {
	a.formatCombo.SetActive(0)
	a.facCombo.SetActive(int(a.state.Facility))
	a.sevCombo.SetActive(int(a.state.Severity))
	a.msgBuf.SetText("")
}
