package processor

import (
	"log"
	"os"
	"time"

	"github.com/yllada/Send-Log-TCP/schema"
)

func SetDefaultSyslogConfig(config *schema.SyslogConfig) {
	// Protocolo, Address, Port y Messages son requeridos, no necesitan valores por defecto

	// Si Hostname está vacío, intentar obtener el nombre del host desde el sistema operativo
	if config.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Printf("Error retrieving hostname: %v", err)
			config.Hostname = "unknown-host" // Valor por defecto si no se puede obtener
		} else {
			config.Hostname = hostname
		}
	}

	// Si Facility es 0 (no se proporcionó), asignar el valor por defecto 1
	if config.Facility == 0 {
		config.Facility = 1
	}

	// Si Severity es 0 (no se proporcionó), asignar el valor por defecto 5 (informativo)
	if config.Severity == 0 {
		config.Severity = 5
	}

	// Si Interval es 0, asignar un valor por defecto de 1 segundo
	if config.Interval == 0 {
		config.Interval = 1 * time.Second
	}
}
