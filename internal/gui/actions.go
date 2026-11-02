//go:build linux && cgo

package gui

import (
	"context"
	"fmt"
	"strings"

	"Sendlog-Syslog/internal/logimport"
	"Sendlog-Syslog/internal/syslog"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// readMessages collects the message textview content, one trimmed line each.
func (a *App) readMessages() []string {
	start, end := a.msgBuf.StartIter(), a.msgBuf.EndIter()
	text := a.msgBuf.Text(start, end, false)
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
}

func (a *App) toast(title string) {
	t := adw.NewToast(title)
	t.SetPriority(adw.ToastPriorityHigh)
	overlay := a.win.Child().(*adw.ToastOverlay)
	overlay.AddToast(t)
}

func (a *App) setStatus(icon, cssClass, text string) {
	a.statusIcon.SetText(icon)
	a.statusIcon.RemoveCSSClass("sendlog-status-ok")
	a.statusIcon.RemoveCSSClass("sendlog-status-err")
	a.statusIcon.RemoveCSSClass("sendlog-status-warn")
	if cssClass != "" {
		a.statusIcon.AddCSSClass(cssClass)
	}
	a.statusText.SetText(text)
}

func (a *App) log(line string) {
	buf := a.details.Buffer()
	start, end := buf.StartIter(), buf.EndIter()
	buf.SetText(buf.Text(start, end, false) + line + "\n")
}

// onToggleConnect opens the live connection (when the button is toggled on)
// or tears it down (when toggled off).
func (a *App) onToggleConnect() {
	if a.toggleGuard {
		return
	}
	if a.connToggle.Active() {
		a.connect()
	} else {
		a.disconnect()
	}
}

// setToggle flips the connect toggle without re-entering onToggleConnect.
func (a *App) setToggle(active bool) {
	a.toggleGuard = true
	a.connToggle.SetActive(active)
	a.toggleGuard = false
}

func (a *App) transportLabel(cfg syslog.Config) string {
	proto := cfg.Protocol
	if cfg.UseTLS {
		proto += "+tls"
	}
	return proto
}

func (a *App) connect() {
	if strings.TrimSpace(a.state.Address) == "" {
		a.setToggle(false)
		a.connToggle.SetLabel("Connect")
		a.setStatus("!", "sendlog-status-err", "Address is required")
		return
	}
	cfg := a.state.ToSyslogConfig()
	a.connToggle.SetSensitive(false)
	a.setStatus("…", "", "Connecting…")
	go func() {
		client, err := syslog.DialClient(cfg)
		glib.IdleAdd(func() {
			a.connToggle.SetSensitive(true)
			if err != nil {
				a.connToggle.SetLabel("Connect")
				a.setToggle(false)
				a.setStatus("✗", "sendlog-status-err", "Connection failed")
				a.log(fmt.Sprintf("connect error: %v", err))
				a.toast("Connection failed")
				return
			}
			a.client = client
			a.connToggle.SetLabel("Disconnect")
			a.setToggle(true)
			label := fmt.Sprintf("Connected (%s) to %s:%s",
				a.transportLabel(cfg), cfg.Address, cfg.Port)
			a.setStatus("✓", "sendlog-status-ok", label)
			a.log("connection ok: " + label)
			a.setConnected(true)
			a.toast("Connected")
		})
	}()
}

func (a *App) disconnect() {
	saidByEngine := false
	if a.engine.IsContinuousRunning() {
		// Closing the socket below stops the engine on its next write; the
		// stats ticker then restores the action buttons and shows the summary.
		a.engine.Stop()
		saidByEngine = true
	}
	if a.client != nil {
		_ = a.client.Close()
		a.client = nil
	}
	a.setConnected(false)
	a.connToggle.SetLabel("Connect")
	a.setToggle(false)
	if !saidByEngine {
		a.setStatus("■", "sendlog-status-warn", "Disconnected")
		a.log("connection closed")
	}
}

// setConnected toggles sensitivity of the controls that need a live connection.
func (a *App) setConnected(connected bool) {
	a.connected = connected
	a.sendBtn.SetSensitive(connected)
	a.stressBtn.SetSensitive(connected)
}

// onSend sends the current message list with progress on the status bar.
func (a *App) onSend() {
	msgs := a.readMessages()
	if len(msgs) == 0 {
		a.setStatus("!", "sendlog-status-err", "No messages to send")
		return
	}
	if !a.connected {
		a.setStatus("!", "sendlog-status-warn", "Connect first")
		return
	}

	a.state.Messages = msgs
	cfg := a.state.ToSyslogConfig()
	a.sendBtn.SetSensitive(false)
	a.statusBar.SetFraction(0)
	a.setStatus("…", "", fmt.Sprintf("Sending %d message(s)…", len(msgs)))
	go func() {
		var resp syslog.Response
		onProgress := func(current, total int, percent float64) {
			frac := percent / 100
			glib.IdleAdd(func() { a.statusBar.SetFraction(frac) })
		}
		if a.client != nil {
			resp = a.client.SendMessages(cfg, msgs, onProgress)
		} else {
			resp = syslog.SendSyslogMessages(cfg, onProgress)
		}
		glib.IdleAdd(func() {
			a.sendBtn.SetSensitive(true)
			a.statusBar.SetFraction(1)
			if len(resp.Errors) > 0 {
				a.setStatus("✗", "sendlog-status-err",
					fmt.Sprintf("Sent %d, %d error(s)", len(resp.SentMessages), len(resp.Errors)))
				for _, e := range resp.Errors {
					a.log(e)
				}
				a.toast("Send finished with errors")
			} else {
				a.setStatus("✓", "sendlog-status-ok",
					fmt.Sprintf("Sent %d message(s)", len(resp.SentMessages)))
				a.log(fmt.Sprintf("sent %d message(s) successfully", len(resp.SentMessages)))
				a.toast(fmt.Sprintf("Sent %d message(s)", len(resp.SentMessages)))
			}
		})
	}()
}

// importFile opens the file chooser and loads the result into the textview.
func (a *App) importFile() {
	dialog := gtk.NewFileDialog()
	dialog.SetTitle("Import log file")
	dialog.SetModal(true)
	dialog.Open(context.Background(), &a.win.Window, func(asyncRes gio.AsyncResulter) {
		file, err := dialog.OpenFinish(asyncRes)
		if err != nil || file == nil {
			glib.IdleAdd(func() {
				if err != nil {
					a.log(fmt.Sprintf("import error: %v", err))
				}
			})
			return
		}
		path := file.Path()
		go func() {
			res, err := logimport.Import(path)
			glib.IdleAdd(func() {
				if err != nil {
					a.setStatus("✗", "sendlog-status-err", "Import failed")
					a.log(fmt.Sprintf("import error: %v", err))
					return
				}
				a.msgBuf.SetText(strings.Join(res.Messages, "\n"))
				a.importLabel.SetText(fmt.Sprintf("%s — %d lines, %d errors",
					baseName(path), res.TotalLines, len(res.Errors)))
				a.setStatus("✓", "sendlog-status-ok",
					fmt.Sprintf("Imported %d message(s)", len(res.Messages)))
				a.log(fmt.Sprintf("imported %d lines from %s (%d errors)",
					res.TotalLines, baseName(path), len(res.Errors)))
				a.toast(fmt.Sprintf("Imported %d message(s)", len(res.Messages)))
			})
		}()
	})
}

func baseName(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[i+1:]
		}
	}
	return p
}

// onStartStress begins a continuous stress run over the live connection
// (when connected) or a one-shot dial (when not), and starts the stats ticker.
func (a *App) onStartStress() {
	a.state.StressMessage = a.stressMsgEnt.Text()
	cfg := a.state.ToStressConfig()
	if a.client != nil {
		if conn, err := a.client.Connection(); err == nil {
			cfg.Conn = conn
		}
	}
	if err := a.engine.Start(cfg, nil); err != nil {
		a.setStatus("✗", "sendlog-status-err", "Stress failed to start")
		a.log(fmt.Sprintf("stress start error: %v", err))
		a.toast("Stress failed")
		return
	}

	a.stressBtn.SetSensitive(false)
	a.stopBtn.SetSensitive(true)
	mode := fmt.Sprintf("%d msg/s, duration=%ds, max=%d",
		cfg.MessagesPerSec, cfg.Duration, cfg.MaxMessages)
	if a.client != nil {
		mode += " (connected)"
	}
	a.setStatus("▶", "sendlog-status-ok", "Stress running")
	a.log("stress started: " + mode)
	a.statsSource = glib.TimeoutAdd(500, func() bool {
		if !a.engine.IsContinuousRunning() {
			a.stopStatsTicker()
			a.stopBtn.SetSensitive(false)
			if a.client != nil {
				a.sendBtn.SetSensitive(true)
				a.stressBtn.SetSensitive(true)
			}
			s := a.engine.GetContinuousStats()
			a.setStatus("■", "sendlog-status-ok",
				fmt.Sprintf("Finished: %d sent, %d errors", s.TotalSent, s.TotalErrors))
			return false
		}
		s := a.engine.GetContinuousStats()
		a.statusText.SetText(fmt.Sprintf("Stress: %d sent, %d errors, %.1f msg/s",
			s.TotalSent, s.TotalErrors, s.CurrentRate))
		if s.Duration > 0 {
			a.statusBar.SetFraction(s.ElapsedSeconds / float64(s.Duration))
		} else {
			a.statusBar.SetFraction(0)
			a.statusBar.Pulse()
		}
		return true
	})
}

// onStop halts the stress run and prints the final totals.
func (a *App) onStop() {
	a.engine.Stop()
	a.stopStatsTicker()
	a.stopBtn.SetSensitive(false)
	if a.client != nil {
		a.sendBtn.SetSensitive(true)
		a.stressBtn.SetSensitive(true)
	}
	stats := a.engine.GetContinuousStats()
	a.setStatus("■", "sendlog-status-warn",
		fmt.Sprintf("Stopped: %d sent, %d errors", stats.TotalSent, stats.TotalErrors))
	a.log(fmt.Sprintf("stress stopped: %d sent, %d errors", stats.TotalSent, stats.TotalErrors))
	a.statusBar.SetFraction(0)
}

func (a *App) stopStatsTicker() {
	if a.statsSource != 0 {
		glib.SourceRemove(a.statsSource)
		a.statsSource = 0
	}
}
