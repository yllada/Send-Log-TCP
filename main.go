package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// Application identifier for single instance lock
const appUniqueID = "com.sendlog-syslog"

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "SendLog Syslog",
		Width:             900,
		Height:            700,
		MinWidth:          900, // Fixed size - disables maximize button
		MinHeight:         700,
		MaxWidth:          900,
		MaxHeight:         700,
		DisableResize:     true, // Prevent resizing completely
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 0, G: 0, B: 0, A: 0}, // Transparent for Mica (Windows only)
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// Prevent accidental file drops in the webview
		DragAndDrop: &options.DragAndDrop{
			DisableWebViewDrop: true,
		},
		// Prevent multiple instances of the application
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: appUniqueID,
		},
		Menu:             nil,
		Logger:           nil,
		LogLevel:         logger.DEBUG,
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options - Mica/Acrylic effect
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableWindowIcon:                 false,
			BackdropType:                      windows.Mica, // Windows 11 Mica effect
			DisableFramelessWindowDecorations: false,
			WebviewUserDataPath:               "",
			ZoomFactor:                        1.0,
			Theme:                             windows.SystemDefault,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "SendLog Syslog",
				Message: "",
				Icon:    icon,
			},
		},
		// Linux platform specific options
		Linux: &linux.Options{
			Icon:                icon,
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
			ProgramName:         "SendLog Syslog",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
