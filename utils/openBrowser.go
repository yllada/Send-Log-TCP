package utils

import (
	"log"
	"os/exec"
	"runtime"
)

// openBrowser abre el navegador predeterminado en la URL proporcionada.
func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		log.Println("unsupported platform")
		return
	}

	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}
