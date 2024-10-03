package schema

import "time"

// Configuración del servidor Syslog
type SyslogConfig struct {
	Address  string
	Protocol string
	Facility int
	Severity int
	Hostname string
	Messages []string
	Interval time.Duration // Enviar logs en intervalos (en segundos)
}
