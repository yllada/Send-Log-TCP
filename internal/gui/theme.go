//go:build linux && cgo

package gui

import _ "embed"

// themeCSS layers a little polish on top of the native Adwaita dark theme.
//
//go:embed theme.css
var themeCSS string
