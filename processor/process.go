package processor

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/yllada/Send-Log-TCP/schema"
)

func BuildSyslogMessage(config schema.SyslogConfig, message string) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	pri := (config.Facility * 8) + config.Severity

	// Añadir CRLF al final del mensaje
	return fmt.Sprintf("%d <%d>1 %s %s %s\r\n", 679, pri, timestamp, config.Hostname, message)
}

func setDefaultSyslogConfig(config *schema.SyslogConfig) {
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

func SendSyslogMessages(config schema.SyslogConfig) {
	// Establecer los valores por defecto si no se proporcionaron
	setDefaultSyslogConfig(&config)
	// Concatenar la dirección con el puerto
	fullAddress := fmt.Sprintf("%s:%s", config.Address, config.Port)

	conn, err := net.Dial(config.Protocol, fullAddress)
	if err != nil {
		log.Printf("Error connecting to %s: %v", fullAddress, err)
		return
	}
	log.Println("Connected to", fullAddress)
	defer conn.Close()

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	for _, message := range config.Messages {
		syslogMessage := BuildSyslogMessage(config, message)
		_, err := conn.Write([]byte(syslogMessage))
		if err != nil {
			log.Printf("Error sending message: %v", err)
		} else {
			log.Printf("Sent message: %s", syslogMessage)
		}
		<-ticker.C
	}
}
