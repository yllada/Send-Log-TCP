package processor

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/yllada/Send-Log-TCP/schema"
)

func BuildSyslogMessage(config schema.SyslogConfig, message string) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	pri := (config.Facility * 8) + config.Severity

	// Añadir CRLF al final del mensaje
	return fmt.Sprintf("%d <%d>1 %s %s %s\r\n", 679, pri, timestamp, config.Hostname, message)
}

func SendSyslogMessages(config schema.SyslogConfig) {
	conn, err := net.Dial(config.Protocol, config.Address)
	if err != nil {
		log.Printf("Error connecting to %s: %v", config.Address, err)
		return
	}
	log.Println("Connected to", config.Address)
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
