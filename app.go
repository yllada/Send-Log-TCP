package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Connection timeouts according to best practices
const (
	connectionTimeout = 10 * time.Second
	writeTimeout      = 30 * time.Second
	readTimeout       = 30 * time.Second
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
	// Clean up any active connections
	a.Disconnect()
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Ensure all connections are closed
	a.Disconnect()
}

// IsConnected returns the current connection state
func (a *App) IsConnected() bool {
	a.connMu.Lock()
	defer a.connMu.Unlock()
	return a.isConnected
}

// dialTLS establece una conexión TLS al servidor remoto
// Si tlsVerify es false, acepta cualquier certificado (útil para certificados autofirmados)
// Si tlsVerify es true, verifica el certificado usando las CA del sistema
func dialTLS(address, port string, tlsVerify bool) (net.Conn, error) {
	// net.JoinHostPort maneja correctamente IPv4 e IPv6
	fullAddress := net.JoinHostPort(address, port)

	// Configuración TLS moderna siguiendo mejores prácticas
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !tlsVerify,
		MinVersion:         tls.VersionTLS12, // TLS 1.2 mínimo (estándar de la industria)
		MaxVersion:         tls.VersionTLS13, // TLS 1.3 máximo (más reciente y seguro)
	}

	// Si se verifica el certificado, configurar el ServerName para SNI
	if tlsVerify {
		tlsConfig.ServerName = address
	}

	// Establecer conexión TLS con timeout
	dialer := &net.Dialer{
		Timeout: connectionTimeout,
	}

	tlsConn, err := tls.DialWithDialer(dialer, "tcp", fullAddress, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}

	// Realizar el handshake TLS explícitamente
	if err := tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	// Log información de la conexión TLS establecida
	state := tlsConn.ConnectionState()
	log.Printf("TLS connection established: version=%s, cipher=%s, server=%s",
		tlsVersionName(state.Version),
		tls.CipherSuiteName(state.CipherSuite),
		state.ServerName,
	)

	return tlsConn, nil
}

// tlsVersionName convierte el código de versión TLS a nombre legible
func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04X)", version)
	}
}

// dialConnection establece una conexión según el protocolo configurado
// Soporta TCP, UDP y TCP+TLS (RFC 5425)
func dialConnection(address, port, protocol string, useTLS, tlsVerify bool) (net.Conn, error) {
	// Si se solicita TLS pero el protocolo es UDP, retornar error
	if useTLS && protocol == "udp" {
		return nil, fmt.Errorf("TLS is not supported over UDP protocol")
	}

	// Si es TCP con TLS
	if protocol == "tcp" && useTLS {
		return dialTLS(address, port, tlsVerify)
	}

	// Conexión estándar (TCP o UDP sin TLS)
	// net.JoinHostPort maneja correctamente IPv4 e IPv6
	fullAddress := net.JoinHostPort(address, port)
	dialer := &net.Dialer{
		Timeout: connectionTimeout,
	}

	conn, err := dialer.Dial(protocol, fullAddress)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return conn, nil
}

// FramingMethod especifica el método de framing para TCP según RFC 6587
// RFC 6587: "Transmission of Syslog Messages over TCP"
// Define dos métodos estándar para delimitar mensajes en streams TCP
type FramingMethod string

const (
	// OctetCounting implementa el método de conteo de octetos (RFC 6587 Section 3.4.1)
	// Formato: MSG-LEN SP SYSLOG-MSG
	// Este es el método RECOMENDADO porque no tiene restricciones sobre el contenido
	OctetCounting FramingMethod = "octet-counting"

	// NonTransparent implementa framing no-transparente (RFC 6587 Section 3.4.2)
	// Formato: SYSLOG-MSG LF
	// Limitación: el mensaje no puede contener el delimitador LF
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
	UseTLS        bool          `json:"UseTLS"`        // true=usar TLS para conexión TCP
	TLSVerify     bool          `json:"TLSVerify"`     // true=verificar certificado del servidor
}

// SyslogResponse to send back to the frontend
type SyslogResponse struct {
	SentMessages []string `json:"sentMessages"`
	Errors       []string `json:"errors"`
}

// CheckConnection verifies if a connection can be established
// Soporta TCP, UDP y TCP+TLS (RFC 5425 para syslog seguro)
func (a *App) CheckConnection(address string, port string, protocol string, useTLS bool, tlsVerify bool) (bool, error) {
	// net.JoinHostPort maneja correctamente IPv4 e IPv6
	fullAddress := net.JoinHostPort(address, port)

	// Establecer conexión con soporte TLS
	conn, err := dialConnection(address, port, protocol, useTLS, tlsVerify)
	if err != nil {
		log.Printf("Error connecting to %s: %v", fullAddress, err)
		return false, err
	}

	// Almacenar la conexión activa
	a.connMu.Lock()
	a.conn = conn
	a.isConnected = true
	a.connMu.Unlock()

	protocolInfo := protocol
	if useTLS {
		protocolInfo = fmt.Sprintf("%s+TLS", protocol)
	}
	log.Printf("Connected to %s via %s", fullAddress, protocolInfo)
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

	// Emit start event
	runtime.EventsEmit(a.ctx, "syslog:sending", map[string]interface{}{
		"total": len(config.Messages),
	})

	// net.JoinHostPort maneja correctamente IPv4 e IPv6
	fullAddress := net.JoinHostPort(config.Address, config.Port)

	// Establecer conexión con soporte TLS
	conn, err := dialConnection(config.Address, config.Port, config.Protocol, config.UseTLS, config.TLSVerify)
	if err != nil {
		log.Printf("Error connecting to %s: %v", fullAddress, err)
		response.Errors = append(response.Errors, fmt.Sprintf("Error connecting to %s: %v", fullAddress, err))
		return response
	}

	protocolInfo := config.Protocol
	if config.UseTLS {
		protocolInfo = fmt.Sprintf("%s+TLS", config.Protocol)
	}
	log.Printf("Connected to %s via %s", fullAddress, protocolInfo)
	defer conn.Close()

	// Enviar mensajes según el protocolo
	if config.Protocol == "tcp" {
		return a.sendTCPMessages(conn, config)
	}
	return a.sendUDPMessages(conn, config)
}

// sendTCPMessages envía mensajes TCP con el framing adecuado según RFC 6587
// Utiliza el módulo Framer para aplicar el método de framing configurado
func (a *App) sendTCPMessages(conn net.Conn, config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	// Crear framer con la configuración adecuada
	framer := NewFramer(FramingConfig{
		Method:           config.FramingMethod,
		ValidateUTF8:     true, // RFC 5424 requiere UTF-8
		MaxMessageLength: 0,    // Sin límite (el servidor puede imponer límites)
	})

	totalMessages := len(config.Messages)
	for i, message := range config.Messages {
		// Construir mensaje syslog según RFC 5424 o RFC 3164
		syslogMsg, err := buildSyslogMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		// Aplicar framing TCP según RFC 6587 usando el módulo dedicado
		framedMsg, err := framer.Frame(syslogMsg)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error framing message: %v", err))
			continue
		}

		// Set write deadline to prevent hanging
		if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			log.Printf("Warning: failed to set write deadline: %v", err)
		}

		// Enviar con manejo robusto de escrituras parciales
		if err := writeAll(conn, framedMsg); err != nil {
			log.Printf("Error sending message: %v", err)
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			log.Printf("Sent message: %s", syslogMsg)
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}

		// Emit progress event
		runtime.EventsEmit(a.ctx, "syslog:progress", map[string]interface{}{
			"current": i + 1,
			"total":   totalMessages,
			"percent": float64(i+1) / float64(totalMessages) * 100,
		})
	}

	// Emit completion event
	runtime.EventsEmit(a.ctx, "syslog:complete", map[string]interface{}{
		"sent":   len(response.SentMessages),
		"errors": len(response.Errors),
	})

	return response
}

// sendUDPMessages envía mensajes UDP (un mensaje por paquete, sin framing)
func (a *App) sendUDPMessages(conn net.Conn, config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	totalMessages := len(config.Messages)
	for i, message := range config.Messages {
		syslogMsg, err := buildSyslogMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		// Set write deadline to prevent hanging
		if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			log.Printf("Warning: failed to set write deadline: %v", err)
		}

		// UDP no requiere framing, un mensaje por paquete
		if err := writeAll(conn, []byte(syslogMsg)); err != nil {
			log.Printf("Error sending message: %v", err)
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			log.Printf("Sent message: %s", syslogMsg)
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}

		// Emit progress event
		runtime.EventsEmit(a.ctx, "syslog:progress", map[string]interface{}{
			"current": i + 1,
			"total":   totalMessages,
			"percent": float64(i+1) / float64(totalMessages) * 100,
		})
	}

	// Emit completion event
	runtime.EventsEmit(a.ctx, "syslog:complete", map[string]interface{}{
		"sent":   len(response.SentMessages),
		"errors": len(response.Errors),
	})

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

// validateConfig valida y normaliza la configuración antes de enviar mensajes
// Aplica valores por defecto según las recomendaciones de los RFCs
func validateConfig(config *SyslogConfig) error {
	// Validaciones requeridas
	if config.Address == "" {
		return fmt.Errorf("address is required")
	}
	if config.Port == "" {
		return fmt.Errorf("port is required")
	}

	// Validar rangos según RFC 5424
	if config.Facility > 23 {
		return fmt.Errorf("facility must be 0-23 (got %d), see RFC 5424 Section 6.2.1", config.Facility)
	}
	if config.Severity > 7 {
		return fmt.Errorf("severity must be 0-7 (got %d), see RFC 5424 Section 6.2.1", config.Severity)
	}

	// Aplicar valores por defecto para framing
	if config.FramingMethod == "" {
		if config.Protocol == "tcp" {
			// RFC 6587 recomienda octet-counting como método preferido
			config.FramingMethod = RecommendedFramingMethod()
			log.Printf("Using recommended framing method: %s", config.FramingMethod)
		}
	}

	// Validar método de framing si se especificó
	if config.FramingMethod != "" && !IsValidFramingMethod(config.FramingMethod) {
		return fmt.Errorf("invalid framing method '%s', must be 'octet-counting' or 'non-transparent'",
			config.FramingMethod)
	}

	// Aplicar valores por defecto para hostname
	if config.Hostname == "" {
		hostname, err := os.Hostname()
		if err == nil && hostname != "" {
			config.Hostname = hostname
		} else {
			config.Hostname = "localhost" // Fallback seguro
		}
	}

	// Aplicar valor por defecto para appname
	if config.Appname == "" {
		config.Appname = "sendlog"
	}

	return nil
}
