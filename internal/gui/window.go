//go:build linux && cgo

package gui

import (
	"Sendlog-Syslog/internal/stress"
	"Sendlog-Syslog/internal/syslog"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const appID = "com.sendlog.gui"

// App owns the GTK application and the main window plus every widget it
// interacts with, so the action handlers can read and write the UI state.
type App struct {
	app   *adw.Application
	win   *gtk.ApplicationWindow
	state *State

	engine *stress.Engine

	connEntry  *gtk.Entry
	portEntry  *gtk.Entry
	protoCombo *gtk.ComboBoxText
	frameCombo *gtk.ComboBoxText
	tlsSwitch  *gtk.Switch
	verifySw   *gtk.Switch
	caEntry    *gtk.Entry
	certEntry  *gtk.Entry
	keyEntry   *gtk.Entry

	formatCombo *gtk.ComboBoxText
	facCombo    *gtk.ComboBoxText
	sevCombo    *gtk.ComboBoxText
	msgView     *gtk.TextView
	msgBuf      *gtk.TextBuffer

	importLabel  *gtk.Label
	stressMsgEnt *gtk.Entry
	rateSpin     *gtk.SpinButton
	durSpin      *gtk.SpinButton
	maxSpin      *gtk.SpinButton
	randSwitch   *gtk.Switch

	connToggle *gtk.ToggleButton
	sendBtn    *gtk.Button
	stressBtn  *gtk.Button
	stopBtn    *gtk.Button

	statusIcon *gtk.Label
	statusText *gtk.Label
	statusBar  *gtk.ProgressBar
	details    *gtk.TextView

	client      *syslog.Client
	connected   bool
	toggleGuard bool
	statsSource glib.SourceHandle
}

// Run starts the GTK main loop and blocks until the window is closed.
func Run() {
	adw.Init()
	registerAppIcon()
	a := &App{state: DefaultState()}
	a.engine = stress.NewEngine()
	a.app = adw.NewApplication(appID, gio.ApplicationFlagsNone)
	a.app.ConnectActivate(func() { a.buildWindow() })
	a.app.ConnectShutdown(func() {
		a.engine.Stop()
		if a.client != nil {
			_ = a.client.Close()
		}
	})
	a.app.Run(nil)
}

func (a *App) buildWindow() {
	adw.StyleManagerGetDefault().SetColorScheme(adw.ColorSchemePreferDark)
	a.loadTheme()

	a.win = gtk.NewApplicationWindow(&a.app.Application)
	a.win.SetDefaultSize(940, 720)

	hb := adw.NewHeaderBar()
	logo := gtk.NewImageFromIconName(appIconName)
	logo.SetPixelSize(32)
	hb.PackStart(logo)
	hb.SetTitleWidget(adw.NewWindowTitle("SendLog", "syslog sender"))
	a.win.SetTitlebar(hb)
	a.win.SetIconName(appIconName)

	overlay := adw.NewToastOverlay()
	a.win.SetChild(overlay)

	overlay.SetChild(a.buildContent())
	a.win.Present()
}

func (a *App) loadTheme() {
	provider := gtk.NewCSSProvider()
	provider.LoadFromString(themeCSS)
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		provider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)
}

func (a *App) buildContent() gtk.Widgetter {
	outer := gtk.NewScrolledWindow()

	box := gtk.NewBox(gtk.OrientationVertical, 12)
	box.SetMarginTop(16)
	box.SetMarginBottom(16)
	box.SetMarginStart(16)
	box.SetMarginEnd(16)

	box.Append(a.buildConnectionFrame())
	box.Append(a.buildMessageFrame())
	box.Append(a.buildStressFrame())
	box.Append(a.buildStatusArea())

	outer.SetChild(box)
	return outer
}
