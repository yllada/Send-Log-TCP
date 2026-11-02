//go:build linux && cgo

package gui

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed icon.png
var iconPNG []byte

// appIconName is the icon theme name the window and .desktop file share.
const appIconName = "sendlog-app"

// registerAppIcon installs the embedded app icon into a temp directory that
// is added to the GTK icon theme for this session.
func registerAppIcon() {
	dir, err := os.MkdirTemp("", "sendlog-icon-")
	if err != nil {
		return
	}
	if err := os.WriteFile(filepath.Join(dir, appIconName+".png"), iconPNG, 0o644); err != nil {
		return
	}
	theme := gtk.IconThemeGetForDisplay(gdk.DisplayGetDefault())
	theme.AddSearchPath(dir)
}
