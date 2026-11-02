//go:build linux && cgo

package gui

// Available reports whether the GTK4 GUI can run here (compiled in).
func Available() bool { return true }
