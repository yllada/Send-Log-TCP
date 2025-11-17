package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

// App struct
type App struct {
	ctx         context.Context
	conn        net.Conn
	connMu      sync.Mutex // Mutex para manejar el acceso a la conexión
	isConnected bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {}

// beforeClose is called when the application is about to quit.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {}

// FramingMethod especifica el método de framing para TCP según RFC 6587
type FramingMethod string

const (
	// OctetCounting usa el método de conteo de octetos (MSG-LEN SP SYSLOG-MSG)
	OctetCounting FramingMethod = "octet-counting"
	// NonTransparent usa delimitador al final (SYSLOG-MSG LF)
	NonTransparent FramingMethod = "non-transparent"
)

type SyslogConfig struct {
	Address       string        `json:"Address"`
	Port          string        `json:"Port"`
	Protocol      string        `json:"Protocol"`
	Messages      []string      `json:"Messages"`
	FramingMethod FramingMethod `json:"FramingMethod"` // Método de framing (solo TCP)
	Facility      uint8         `json:"Facility"`      // 0-23 (ej: 16=local0)
	Severity      uint8         `json:"Severity"`      // 0-7 (ej: 6=info)
	Hostname      string        `json:"Hostname"`      // Hostname del sistema
	Appname       string        `json:"Appname"`       // Nombre de la aplicación
	UseRFC5424    bool          `json:"UseRFC5424"`    // true=RFC5424, false=RFC3164
}

// SyslogResponse to send back to the frontend
type SyslogResponse struct {
	SentMessages []string `json:"sentMessages"`
	Errors       []string `json:"errors"`
}

// CheckConnection verifies if a connection can be established
func (a *App) CheckConnection(address string, port string, protocol string) (bool, error) {
	fullAddress := fmt.Sprintf("%s:%s", address, port)
	conn, err := net.Dial(protocol, fullAddress)
	if err != nil {
		log.Printf("Error connecting to %s: %v", fullAddress, err)
		return false, err
	}
	// Almacenar la conexión activa
	a.connMu.Lock()
	a.conn = conn
	a.isConnected = true
	a.connMu.Unlock()

	log.Println("Connected to", fullAddress)
	return true, nil
}

// Disconnect cierra la conexión activa si existe
func (a *App) Disconnect() {
	a.connMu.Lock()
	defer a.connMu.Unlock()

	if a.isConnected {
		err := a.conn.Close()
		if err != nil {
			log.Printf("Error disconnecting: %v", err)
		} else {
			log.Println("Disconnected successfully.")
		}
		a.isConnected = false
	}
}

// SendSyslogMessages sends syslog messages and returns the result
func (a *App) SendSyslogMessages(config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	// Validar configuración
	if err := validateConfig(&config); err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Invalid configuration: %v", err))
		return response
	}

	fullAddress := fmt.Sprintf("%s:%s", config.Address, config.Port)

	conn, err := net.Dial(config.Protocol, fullAddress)
	if err != nil {
		log.Printf("Error connecting to %s: %v", fullAddress, err)
		response.Errors = append(response.Errors, fmt.Sprintf("Error connecting to %s: %v", fullAddress, err))
		return response
	}
	log.Println("Connected to", fullAddress)
	defer conn.Close()

	// Enviar mensajes según el protocolo
	if config.Protocol == "tcp" {
		return sendTCPMessages(conn, config)
	}
	return sendUDPMessages(conn, config)
}

// sendTCPMessages envía mensajes TCP con el framing adecuado
func sendTCPMessages(conn net.Conn, config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	for _, message := range config.Messages {
		syslogMsg, err := buildSyslogMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		// Aplicar framing según el método configurado
		framedMsg, err := applyFraming(syslogMsg, config.FramingMethod)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error framing message: %v", err))
			continue
		}

		// Enviar con manejo de escrituras parciales
		if err := writeAll(conn, framedMsg); err != nil {
			log.Printf("Error sending message: %v", err)
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			log.Printf("Sent message: %s", syslogMsg)
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}
	}

	return response
}

// sendUDPMessages envía mensajes UDP (un mensaje por paquete, sin framing)
func sendUDPMessages(conn net.Conn, config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	for _, message := range config.Messages {
		syslogMsg, err := buildSyslogMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		// UDP no requiere framing, un mensaje por paquete
		if err := writeAll(conn, []byte(syslogMsg)); err != nil {
			log.Printf("Error sending message: %v", err)
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			log.Printf("Sent message: %s", syslogMsg)
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}
	}

	return response
}

// buildSyslogMessage construye un mensaje syslog válido según RFC 5424 o RFC 3164
func buildSyslogMessage(config SyslogConfig, message string) (string, error) {
	// Calcular prioridad: PRI = Facility * 8 + Severity
	priority := config.Facility*8 + config.Severity

	// Validar rango de prioridad
	if priority > 191 {
		return "", fmt.Errorf("invalid priority %d (facility=%d, severity=%d)", priority, config.Facility, config.Severity)
	}

	if config.UseRFC5424 {
		return buildRFC5424Message(priority, config, message), nil
	}
	return buildRFC3164Message(priority, config, message), nil
}

// buildRFC5424Message construye mensaje según RFC 5424
// Formato: <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID STRUCTURED-DATA MSG
func buildRFC5424Message(priority uint8, config SyslogConfig, message string) string {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")

	hostname := config.Hostname
	if hostname == "" {
		hostname = "-"
	}

	appname := config.Appname
	if appname == "" {
		appname = "-"
	}

	// RFC 5424: <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID SD MSG
	return fmt.Sprintf("<%d>1 %s %s %s - - - %s",
		priority, timestamp, hostname, appname, message)
}

// buildRFC3164Message construye mensaje según RFC 3164 (BSD syslog)
// Formato: <PRI>TIMESTAMP HOSTNAME TAG: MSG
func buildRFC3164Message(priority uint8, config SyslogConfig, message string) string {
	// RFC 3164 usa formato "Jan  2 15:04:05" (mes en inglés, día con padding de espacio)
	timestamp := time.Now().Format("Jan  2 15:04:05")

	hostname := config.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
		if hostname == "" {
			hostname = "localhost"
		}
	}

	appname := config.Appname
	if appname == "" {
		appname = "app"
	}

	// RFC 3164: <PRI>TIMESTAMP HOSTNAME TAG: MSG
	return fmt.Sprintf("<%d>%s %s %s: %s",
		priority, timestamp, hostname, appname, message)
}

// applyFraming aplica el método de framing correspondiente
func applyFraming(syslogMsg string, method FramingMethod) ([]byte, error) {
	msgBytes := []byte(syslogMsg)

	switch method {
	case OctetCounting:
		// RFC 6587 Octet Counting: MSG-LEN SP SYSLOG-MSG
		// MSG-LEN es el número de octetos de SYSLOG-MSG
		msgLen := len(msgBytes)
		framedMsg := fmt.Sprintf("%d %s", msgLen, syslogMsg)
		return []byte(framedMsg), nil

	case NonTransparent:
		// RFC 6587 Non-Transparent-Framing: SYSLOG-MSG LF
		framedMsg := append(msgBytes, '\n')
		return framedMsg, nil

	default:
		// Por defecto, usar non-transparent framing
		framedMsg := append(msgBytes, '\n')
		return framedMsg, nil
	}
}

// writeAll garantiza que todos los bytes se escriban (maneja escrituras parciales)
func writeAll(w io.Writer, data []byte) error {
	totalWritten := 0
	dataLen := len(data)

	for totalWritten < dataLen {
		n, err := w.Write(data[totalWritten:])
		if err != nil {
			return fmt.Errorf("write failed after %d/%d bytes: %w", totalWritten, dataLen, err)
		}
		totalWritten += n
	}

	return nil
}

// validateConfig valida la configuración antes de enviar
func validateConfig(config *SyslogConfig) error {
	if config.Address == "" {
		return fmt.Errorf("address is required")
	}
	if config.Port == "" {
		return fmt.Errorf("port is required")
	}
	if config.Facility > 23 {
		return fmt.Errorf("facility must be 0-23, got %d", config.Facility)
	}
	if config.Severity > 7 {
		return fmt.Errorf("severity must be 0-7, got %d", config.Severity)
	}

	// Valores por defecto
	if config.FramingMethod == "" {
		if config.Protocol == "tcp" {
			config.FramingMethod = OctetCounting // Por defecto usar octet-counting en TCP
		}
	}

	if config.Hostname == "" {
		hostname, _ := os.Hostname()
		if hostname != "" {
			config.Hostname = hostname
		}
	}

	if config.Appname == "" {
		config.Appname = "sendlog"
	}

	return nil
}
