package processor

import (
	"fmt"
	"log"
	"net"
	"time"
)

// buildSyslogMessage construye el mensaje Syslog en el formato especificado
func BuildSyslogMessage(config SyslogConfig, message string) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	pri := (config.Facility * 8) + config.Severity

	// Formato del mensaje: "<PRI>1 TIMESTAMP HOSTNAME MESSAGE"
	return fmt.Sprintf("%d <%d>1 %s %s %s", 679, pri, timestamp, config.Hostname, message)
}

// sendSyslogMessage envía el mensaje Syslog usando TCP o UDP
func SendSyslogMessage(config SyslogConfig, message string) {
	conn, err := net.Dial(config.Protocol, config.Address)
	if err != nil {
		log.Fatalf("Error connecting to %s: %v", config.Address, err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		log.Printf("Error sending message: %v", err)
	} else {
		log.Printf("Sent message: %s", message)
	}
}
