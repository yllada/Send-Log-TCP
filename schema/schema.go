package schema

import "time"

type SyslogConfig struct {
	Address  string
	Port     string
	Protocol string
	Facility int
	Severity int
	Hostname string
	Messages []string
	Interval time.Duration 
}
