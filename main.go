package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/yllada/Send-Log-TCP/processor"
	"github.com/yllada/Send-Log-TCP/schema"
	"github.com/yllada/Send-Log-TCP/utils"
)

// Syslog severities (levels)
const (
	Emergency = 0
	Alert     = 1
	Critical  = 2
	Error     = 3
	Warning   = 4
	Notice    = 5
	Info      = 6
	Debug     = 7
)

// Syslog facilities
const (
	Local0 = 16
	Local1 = 17
	Local2 = 18
	Local3 = 19
	Local4 = 20
	Local5 = 21
	Local6 = 22
	Local7 = 23
)

func main() {
	// Definir los flags para argumentos de línea de comandos
	address := flag.String("address", "127.0.0.1:7018", "Syslog server address")
	protocol := flag.String("protocol", "tcp", "Protocol (udp or tcp)")
	facility := flag.Int("facility", Local0, "Syslog facility")
	severity := flag.Int("severity", Info, "Syslog severity")
	hostname := flag.String("hostname", "", "Hostname")
	messages := flag.String("messages", "", "Comma-separated log messages")
	interval := flag.Duration("interval", 5*time.Second, "Interval for sending logs (e.g., 5s)")

	// Añadir ayuda
	flag.Usage = func() {
		fmt.Println("Usage: syslog_sender [options]")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validar el protocolo
	if *protocol != "tcp" && *protocol != "udp" {
		log.Fatalf("Invalid protocol: %s (must be tcp or udp)", *protocol)
	}

	// Validar dirección
	if !utils.IsValidAddress(*address) {
		log.Fatalf("Invalid address: %s (must be in the format IP:Port)", *address)
	}

	// Si el hostname no está definido, obtenemos el nombre del sistema
	if *hostname == "" {
		hostnameStr, err := os.Hostname()
		if err != nil {
			log.Fatalf("Error getting hostname: %v", err)
		}
		*hostname = hostnameStr
	}

	// Separar los mensajes en un slice
	msgs := strings.Split(*messages, ",")
	if len(msgs) == 0 || (len(msgs) == 1 && msgs[0] == "") {
		log.Fatalf("At least one log message must be provided")
	}

	config := schema.SyslogConfig{
		Address:  *address,
		Protocol: *protocol,
		Facility: *facility,
		Severity: *severity,
		Hostname: *hostname,
		Messages: msgs,
		Interval: *interval,
	}

	processor.SendSyslogMessages(config)
}
