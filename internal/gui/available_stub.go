//go:build !linux || !cgo

package gui

// Available reports whether the GTK4 GUI can run here. The GUI is only
// compiled for linux with CGO; everywhere else callers fall back to the TUI.
func Available() bool { return false }

// Run is a no-op on platforms without the GTK4 GUI. It exists so the binary
// compiles uniformly and is never called when Available() is false.
func Run() {
	panic("sendlog gui requires linux with CGO and GTK4")
}
