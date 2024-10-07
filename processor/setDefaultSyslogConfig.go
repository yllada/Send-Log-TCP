package processor

import (
	"log"
	"os"
	"time"

	"github.com/yllada/Send-Log-TCP/schema"
)

func SetDefaultSyslogConfig(config *schema.SyslogConfig) {

	if config.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Printf("Error retrieving hostname: %v", err)
			config.Hostname = "unknown-host" 
		} else {
			config.Hostname = hostname
		}
	}

	if config.Facility == 0 {
		config.Facility = 1
	}

	if config.Severity == 0 {
		config.Severity = 5
	}

	if config.Interval == 0 {
		config.Interval = 1 * time.Second
	}
}
