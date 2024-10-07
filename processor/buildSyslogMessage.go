package processor

import (
	"fmt"
	"time"

	"github.com/yllada/Send-Log-TCP/schema"
)

func BuildSyslogMessage(config schema.SyslogConfig, message string) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	pri := (config.Facility * 8) + config.Severity

	return fmt.Sprintf("%d <%d>1 %s %s %s\r\n", 679, pri, timestamp, config.Hostname, message)
}
