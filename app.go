package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Version information - injected at build time via ldflags
// Use: go build -ldflags "-X main.Version=1.4.2"
var Version = "dev"

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

	// Continuous/Stress test fields
	continuousCancel    context.CancelFunc
	continuousMu        sync.RWMutex
	continuousStats     ContinuousStats
	isContinuousRunning bool
}

// ContinuousSendConfig configuration for continuous/stress test mode
type ContinuousSendConfig struct {
	// Connection settings (inherited from SyslogConfig)
	Address        string        `json:"Address"`
	Port           string        `json:"Port"`
	Protocol       string        `json:"Protocol"`
	Message        string        `json:"Message"` // Single message template
	FramingMethod  FramingMethod `json:"FramingMethod"`
	Facility       uint8         `json:"Facility"`
	Severity       uint8         `json:"Severity"`
	Hostname       string        `json:"Hostname"`
	Appname        string        `json:"Appname"`
	UseRFC5424     bool          `json:"UseRFC5424"`
	UseTLS         bool          `json:"UseTLS"`
	TLSVerify      bool          `json:"TLSVerify"`
	CACertPath     string        `json:"CACertPath"`
	ClientCertPath string        `json:"ClientCertPath"`
	ClientKeyPath  string        `json:"ClientKeyPath"`

	// Continuous send settings
	Duration       int  `json:"Duration"`       // Duration in seconds (0 = indefinite)
	MessagesPerSec int  `json:"MessagesPerSec"` // Rate: messages per second
	MaxMessages    int  `json:"MaxMessages"`    // Total limit (0 = no limit)
	RandomizeData  bool `json:"RandomizeData"`  // Add sequence number and timestamp
}

// ContinuousStats real-time statistics for continuous send
type ContinuousStats struct {
	TotalSent      int64   `json:"totalSent"`
	TotalErrors    int64   `json:"totalErrors"`
	CurrentRate    float64 `json:"currentRate"` // Actual msgs/sec
	ElapsedSeconds float64 `json:"elapsedSeconds"`
	IsRunning      bool    `json:"isRunning"`
	StartTime      int64   `json:"startTime"`  // Unix timestamp
	TargetRate     int     `json:"targetRate"` // Configured rate
	Duration       int     `json:"duration"`   // Configured duration
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
func (a *App) domReady(ctx context.Context) {
	// Center the window on the screen
	runtime.WindowCenter(a.ctx)
}

// beforeClose is called when the application is about to quit.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	// If there's an active connection, ask for confirmation
	a.connMu.Lock()
	isConnected := a.isConnected
	a.connMu.Unlock()

	if isConnected {
		result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:          runtime.QuestionDialog,
			Title:         "Confirm Exit",
			Message:       "There is an active connection. Are you sure you want to exit?",
			Buttons:       []string{"Yes", "No"},
			DefaultButton: "No",
			CancelButton:  "No",
		})
		if err != nil || result != "Yes" {
			return true // Prevent close
		}
	}

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

// GetVersion returns the application version
// The version is injected at build time via ldflags
func (a *App) GetVersion() string {
	return Version
}

// OpenURL opens a URL in the default system browser
func (a *App) OpenURL(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
	runtime.LogDebug(a.ctx, "Opened URL in browser: "+url)
}

// OpenGitHub opens the project's GitHub repository in the default browser
func (a *App) OpenGitHub() {
	runtime.BrowserOpenURL(a.ctx, "https://github.com/yllada/Send-Log-TCP")
}

// dialTLS establece una conexión TLS al servidor remoto
// Si tlsVerify es false, acepta cualquier certificado (útil para certificados autofirmados)
// Si tlsVerify es true, verifica el certificado usando las CA del sistema o una CA personalizada
// caCertPath permite especificar un certificado CA personalizado para verificación
// clientCertPath y clientKeyPath permiten autenticación mutua TLS (mTLS)
func dialTLS(address, port string, tlsVerify bool, caCertPath, clientCertPath, clientKeyPath string) (net.Conn, error) {
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

		// Si se proporciona un certificado CA personalizado, cargarlo
		if caCertPath != "" {
			rootCAs, err := loadCACertificate(caCertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA certificate: %w", err)
			}
			tlsConfig.RootCAs = rootCAs
			log.Printf("Using custom CA certificate: %s", caCertPath)
		}
	}

	// Cargar certificado del cliente para mTLS si se proporcionan ambos archivos
	if clientCertPath != "" && clientKeyPath != "" {
		clientCert, err := loadClientCertificate(clientCertPath, clientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
		log.Printf("Using client certificate for mTLS: %s", clientCertPath)
	} else if clientCertPath != "" || clientKeyPath != "" {
		// Solo uno de los dos fue proporcionado
		return nil, fmt.Errorf("both client certificate and key are required for mTLS")
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

// loadCACertificate carga un certificado CA desde un archivo PEM
// Soporta archivos con uno o más certificados (CA bundle)
func loadCACertificate(certPath string) (*x509.CertPool, error) {
	// Leer el archivo de certificado
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
	}

	// Crear un nuevo pool de certificados
	rootCAs := x509.NewCertPool()

	// Agregar el certificado al pool
	if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to parse CA certificate (invalid PEM format)")
	}

	return rootCAs, nil
}

// loadClientCertificate carga el certificado y la clave privada del cliente para mTLS
func loadClientCertificate(certPath, keyPath string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load client certificate/key pair: %w", err)
	}
	return cert, nil
}

// dialConnection establece una conexión según el protocolo configurado
// Soporta TCP, UDP y TCP+TLS (RFC 5425) con mTLS opcional
func dialConnection(address, port, protocol string, useTLS, tlsVerify bool, caCertPath, clientCertPath, clientKeyPath string) (net.Conn, error) {
	// Si se solicita TLS pero el protocolo es UDP, retornar error
	if useTLS && protocol == "udp" {
		return nil, fmt.Errorf("TLS is not supported over UDP protocol")
	}

	// Si es TCP con TLS
	if protocol == "tcp" && useTLS {
		return dialTLS(address, port, tlsVerify, caCertPath, clientCertPath, clientKeyPath)
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
	Address        string        `json:"Address"`
	Port           string        `json:"Port"`
	Protocol       string        `json:"Protocol"`
	Messages       []string      `json:"Messages"`
	FramingMethod  FramingMethod `json:"FramingMethod"`  // Método de framing (solo TCP)
	Facility       uint8         `json:"Facility"`       // 0-23 (ej: 16=local0)
	Severity       uint8         `json:"Severity"`       // 0-7 (ej: 6=info)
	Hostname       string        `json:"Hostname"`       // Hostname del sistema
	Appname        string        `json:"Appname"`        // Nombre de la aplicación
	UseRFC5424     bool          `json:"UseRFC5424"`     // true=RFC5424, false=RFC3164
	UseTLS         bool          `json:"UseTLS"`         // true=usar TLS para conexión TCP
	TLSVerify      bool          `json:"TLSVerify"`      // true=verificar certificado del servidor
	CACertPath     string        `json:"CACertPath"`     // Ruta al certificado CA personalizado (opcional)
	ClientCertPath string        `json:"ClientCertPath"` // Ruta al certificado del cliente para mTLS (opcional)
	ClientKeyPath  string        `json:"ClientKeyPath"`  // Ruta a la clave privada del cliente para mTLS (opcional)
}

// SyslogResponse to send back to the frontend
type SyslogResponse struct {
	SentMessages []string `json:"sentMessages"`
	Errors       []string `json:"errors"`
}

// SelectCACertificate abre un diálogo para seleccionar un archivo de certificado CA
// Retorna la ruta del archivo seleccionado o una cadena vacía si se cancela
func (a *App) SelectCACertificate() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select CA Certificate",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Certificate Files (*.pem, *.crt, *.cer)",
				Pattern:     "*.pem;*.crt;*.cer;*.ca-bundle",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// SelectClientCertificate abre un diálogo para seleccionar el certificado del cliente (mTLS)
// Retorna la ruta del archivo seleccionado o una cadena vacía si se cancela
func (a *App) SelectClientCertificate() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Client Certificate",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Certificate Files (*.pem, *.crt, *.cer)",
				Pattern:     "*.pem;*.crt;*.cer",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// SelectClientKey abre un diálogo para seleccionar la clave privada del cliente (mTLS)
// Retorna la ruta del archivo seleccionado o una cadena vacía si se cancela
func (a *App) SelectClientKey() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Client Private Key",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Key Files (*.pem, *.key)",
				Pattern:     "*.pem;*.key",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// CheckConnection verifies if a connection can be established
// Soporta TCP, UDP y TCP+TLS (RFC 5425 para syslog seguro)
func (a *App) CheckConnection(address string, port string, protocol string, useTLS bool, tlsVerify bool, caCertPath string, clientCertPath string, clientKeyPath string) (bool, error) {
	// net.JoinHostPort maneja correctamente IPv4 e IPv6
	fullAddress := net.JoinHostPort(address, port)

	// Establecer conexión con soporte TLS
	conn, err := dialConnection(address, port, protocol, useTLS, tlsVerify, caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Error connecting to %s: %v", fullAddress, err))
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
	runtime.LogInfo(a.ctx, fmt.Sprintf("Connected to %s via %s", fullAddress, protocolInfo))
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

	// Establecer conexión con soporte TLS y mTLS
	conn, err := dialConnection(config.Address, config.Port, config.Protocol, config.UseTLS, config.TLSVerify, config.CACertPath, config.ClientCertPath, config.ClientKeyPath)
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
			runtime.LogWarning(a.ctx, fmt.Sprintf("Failed to set write deadline: %v", err))
		}

		// Enviar con manejo robusto de escrituras parciales
		if err := writeAll(conn, framedMsg); err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Error sending message: %v", err))
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			runtime.LogDebug(a.ctx, fmt.Sprintf("Sent message: %s", syslogMsg))
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
			runtime.LogWarning(a.ctx, fmt.Sprintf("Failed to set write deadline: %v", err))
		}

		// UDP no requiere framing, un mensaje por paquete
		if err := writeAll(conn, []byte(syslogMsg)); err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Error sending message: %v", err))
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			runtime.LogDebug(a.ctx, fmt.Sprintf("Sent message: %s", syslogMsg))
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

// ================================================================================
// CONTINUOUS / STRESS TEST MODE
// ================================================================================

// IsContinuousRunning returns whether continuous send is currently active
func (a *App) IsContinuousRunning() bool {
	a.continuousMu.RLock()
	defer a.continuousMu.RUnlock()
	return a.isContinuousRunning
}

// GetContinuousStats returns current statistics for continuous send
func (a *App) GetContinuousStats() ContinuousStats {
	a.continuousMu.RLock()
	defer a.continuousMu.RUnlock()
	return a.continuousStats
}

// StartContinuousSend starts sending messages continuously based on config
// Uses goroutine with context cancellation for clean shutdown
func (a *App) StartContinuousSend(config ContinuousSendConfig) error {
	a.continuousMu.Lock()
	if a.isContinuousRunning {
		a.continuousMu.Unlock()
		return fmt.Errorf("continuous send already running")
	}

	// Create cancellation context
	ctx, cancel := context.WithCancel(a.ctx)
	a.continuousCancel = cancel
	a.isContinuousRunning = true
	a.continuousStats = ContinuousStats{
		TotalSent:   0,
		TotalErrors: 0,
		CurrentRate: 0,
		IsRunning:   true,
		StartTime:   time.Now().Unix(),
		TargetRate:  config.MessagesPerSec,
		Duration:    config.Duration,
	}
	a.continuousMu.Unlock()

	// Emit started event
	runtime.EventsEmit(a.ctx, "continuous:started", a.continuousStats)

	// Start the background goroutine
	go a.runContinuousSend(ctx, config)

	return nil
}

// StopContinuousSend stops the continuous send operation
func (a *App) StopContinuousSend() {
	a.continuousMu.Lock()
	defer a.continuousMu.Unlock()

	if a.continuousCancel != nil {
		a.continuousCancel()
		a.continuousCancel = nil
	}
	a.isContinuousRunning = false
	a.continuousStats.IsRunning = false

	// Emit stopped event
	runtime.EventsEmit(a.ctx, "continuous:stopped", a.continuousStats)
}

// runContinuousSend is the goroutine that handles continuous message sending
func (a *App) runContinuousSend(ctx context.Context, config ContinuousSendConfig) {
	defer func() {
		a.continuousMu.Lock()
		a.isContinuousRunning = false
		a.continuousStats.IsRunning = false
		a.continuousMu.Unlock()
		runtime.EventsEmit(a.ctx, "continuous:stopped", a.GetContinuousStats())
	}()

	// Build syslog config from continuous config
	syslogConfig := SyslogConfig{
		Address:        config.Address,
		Port:           config.Port,
		Protocol:       config.Protocol,
		FramingMethod:  config.FramingMethod,
		Facility:       config.Facility,
		Severity:       config.Severity,
		Hostname:       config.Hostname,
		Appname:        config.Appname,
		UseRFC5424:     config.UseRFC5424,
		UseTLS:         config.UseTLS,
		TLSVerify:      config.TLSVerify,
		CACertPath:     config.CACertPath,
		ClientCertPath: config.ClientCertPath,
		ClientKeyPath:  config.ClientKeyPath,
	}

	// Validate config
	if err := validateConfig(&syslogConfig); err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Invalid config: %v", err))
		runtime.EventsEmit(a.ctx, "continuous:error", err.Error())
		return
	}

	// Establish connection
	conn, err := dialConnection(
		config.Address, config.Port, config.Protocol,
		config.UseTLS, config.TLSVerify,
		config.CACertPath, config.ClientCertPath, config.ClientKeyPath,
	)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Connection failed: %v", err))
		runtime.EventsEmit(a.ctx, "continuous:error", err.Error())
		return
	}
	defer conn.Close()

	// Create framer for TCP
	var framer *Framer
	if config.Protocol == "tcp" {
		framer = NewFramer(FramingConfig{
			Method:           config.FramingMethod,
			ValidateUTF8:     true,
			MaxMessageLength: 0,
		})
	}

	// Calculate interval for rate limiting
	// Using time.Ticker for precise rate control
	interval := time.Second / time.Duration(config.MessagesPerSec)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Duration timer if configured
	var durationTimer <-chan time.Time
	if config.Duration > 0 {
		durationTimer = time.After(time.Duration(config.Duration) * time.Second)
	}

	// Stats tracking
	startTime := time.Now()
	var sentCount, errorCount int64
	var sequenceNum int64
	lastStatsUpdate := time.Now()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			// Cancelled by StopContinuousSend
			return

		case <-durationTimer:
			// Duration expired
			runtime.LogInfo(a.ctx, "Continuous send duration completed")
			return

		case <-ticker.C:
			// Check max messages limit
			if config.MaxMessages > 0 && sentCount >= int64(config.MaxMessages) {
				runtime.LogInfo(a.ctx, fmt.Sprintf("Max messages limit reached: %d", config.MaxMessages))
				return
			}

			// Build message with optional randomization
			message := config.Message
			if config.RandomizeData {
				sequenceNum++
				message = fmt.Sprintf("[seq=%d ts=%s] %s",
					sequenceNum,
					time.Now().Format("15:04:05.000"),
					config.Message,
				)
			}

			// Build syslog message
			syslogMsg, err := buildSyslogMessage(syslogConfig, message)
			if err != nil {
				errorCount++
				continue
			}

			// Send message
			var sendErr error
			if config.Protocol == "tcp" && framer != nil {
				framedMsg, err := framer.Frame(syslogMsg)
				if err != nil {
					errorCount++
					continue
				}
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				sendErr = writeAll(conn, framedMsg)
			} else {
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				sendErr = writeAll(conn, []byte(syslogMsg))
			}

			if sendErr != nil {
				errorCount++
				runtime.LogWarning(a.ctx, fmt.Sprintf("Send error: %v", sendErr))
			} else {
				sentCount++
			}

			// Update stats every 250ms to reduce overhead
			if time.Since(lastStatsUpdate) >= 250*time.Millisecond {
				elapsed := time.Since(startTime).Seconds()
				currentRate := float64(sentCount) / elapsed

				a.continuousMu.Lock()
				a.continuousStats.TotalSent = sentCount
				a.continuousStats.TotalErrors = errorCount
				a.continuousStats.ElapsedSeconds = elapsed
				a.continuousStats.CurrentRate = currentRate
				stats := a.continuousStats
				a.continuousMu.Unlock()

				runtime.EventsEmit(a.ctx, "continuous:stats", stats)
				lastStatsUpdate = time.Now()
			}
		}
	}
}
