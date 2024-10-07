package processor

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/yllada/Send-Log-TCP/schema"
)

func SendSyslogMessages(config schema.SyslogConfig) {
	SetDefaultSyslogConfig(&config)
	
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
