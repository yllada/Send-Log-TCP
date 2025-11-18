package main

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

// FramingConfig contiene la configuración para el framing de mensajes syslog sobre TCP
// según RFC 6587: "Transmission of Syslog Messages over TCP"
type FramingConfig struct {
	Method           FramingMethod // Método de framing a utilizar
	ValidateUTF8     bool          // Si true, valida que el mensaje sea UTF-8 válido (RFC 5424)
	MaxMessageLength int           // Longitud máxima del mensaje en bytes (0 = sin límite)
}

// DefaultFramingConfig retorna una configuración por defecto siguiendo las recomendaciones
// de RFC 6587 (octet-counting es el método preferido)
func DefaultFramingConfig() FramingConfig {
	return FramingConfig{
		Method:           OctetCounting, // RFC 6587 recomienda octet-counting
		ValidateUTF8:     true,          // RFC 5424 requiere UTF-8
		MaxMessageLength: 0,             // Sin límite (el servidor puede imponer sus propios límites)
	}
}

// Framer proporciona funcionalidad de framing para mensajes syslog sobre TCP
// Implementa los métodos definidos en RFC 6587
type Framer struct {
	config FramingConfig
}

// NewFramer crea un nuevo framer con la configuración especificada
func NewFramer(config FramingConfig) *Framer {
	return &Framer{
		config: config,
	}
}

// Frame aplica el método de framing configurado al mensaje syslog
// Retorna el mensaje enmarcado listo para envío sobre TCP o un error si el mensaje no es válido
//
// RFC 6587 define dos métodos estándar:
//   - Octet Counting (Section 3.4.1): Prefija el mensaje con su longitud
//   - Non-Transparent Framing (Section 3.4.2): Agrega un delimitador LF al final
func (f *Framer) Frame(syslogMsg string) ([]byte, error) {
	// Validaciones previas al framing
	if err := f.validateMessage(syslogMsg); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	msgBytes := []byte(syslogMsg)

	switch f.config.Method {
	case OctetCounting:
		return f.applyOctetCounting(msgBytes), nil

	case NonTransparent:
		return f.applyNonTransparent(msgBytes), nil

	default:
		// Por seguridad, usar octet-counting si el método es desconocido
		return f.applyOctetCounting(msgBytes), nil
	}
}

// applyOctetCounting implementa el método de conteo de octetos según RFC 6587 Section 3.4.1
//
// Formato: MSG-LEN SP SYSLOG-MSG
// donde:
//   - MSG-LEN: número de octetos del SYSLOG-MSG (en ASCII decimal)
//   - SP: carácter espacio (0x20)
//   - SYSLOG-MSG: el mensaje syslog completo
//
// Ejemplo: "119 <34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - BOM'su root' failed for lonvick on /dev/pts/8"
//
// Este método es preferible porque:
//   - No requiere escaneo del contenido del mensaje
//   - Permite cualquier byte en el mensaje (incluyendo LF)
//   - Es más eficiente para parsear
func (f *Framer) applyOctetCounting(msgBytes []byte) []byte {
	msgLen := len(msgBytes)

	// Pre-calcular el tamaño exacto del buffer necesario
	// Esto evita realocaciones de memoria
	lenStr := fmt.Sprintf("%d", msgLen)
	totalLen := len(lenStr) + 1 + msgLen // lenStr + SP + mensaje

	// Usar bytes.Buffer para construcción eficiente en memoria
	buf := bytes.NewBuffer(make([]byte, 0, totalLen))
	buf.WriteString(lenStr)
	buf.WriteByte(' ')  // SP (espacio)
	buf.Write(msgBytes) // SYSLOG-MSG

	return buf.Bytes()
}

// applyNonTransparent implementa el método de framing no-transparente según RFC 6587 Section 3.4.2
//
// Formato: SYSLOG-MSG LF
// donde:
//   - SYSLOG-MSG: el mensaje syslog completo
//   - LF: carácter line feed (0x0A, '\n')
//
// Limitaciones importantes:
//   - El mensaje NO puede contener el delimitador LF
//   - Si el mensaje contiene LF, será interpretado como múltiples mensajes
//
// Este método es compatible con implementaciones legacy pero tiene limitaciones
func (f *Framer) applyNonTransparent(msgBytes []byte) []byte {
	// Pre-alocar buffer con capacidad exacta
	buf := bytes.NewBuffer(make([]byte, 0, len(msgBytes)+1))
	buf.Write(msgBytes)
	buf.WriteByte('\n') // LF (line feed)

	return buf.Bytes()
}

// validateMessage valida el mensaje según la configuración antes del framing
func (f *Framer) validateMessage(msg string) error {
	if msg == "" {
		return fmt.Errorf("message cannot be empty")
	}

	// Validar longitud máxima si está configurada
	if f.config.MaxMessageLength > 0 && len(msg) > f.config.MaxMessageLength {
		return fmt.Errorf("message length %d exceeds maximum %d bytes",
			len(msg), f.config.MaxMessageLength)
	}

	// RFC 5424 requiere que los mensajes sean UTF-8 válidos
	if f.config.ValidateUTF8 && !utf8.ValidString(msg) {
		return fmt.Errorf("message contains invalid UTF-8 sequences (RFC 5424 requires valid UTF-8)")
	}

	// Para non-transparent framing, verificar que no contenga LF
	// RFC 6587 Section 3.4.2: el mensaje no puede contener el delimitador
	if f.config.Method == NonTransparent {
		for i := 0; i < len(msg); i++ {
			if msg[i] == '\n' {
				return fmt.Errorf("message contains line feed (LF) which is not allowed in non-transparent framing (use octet-counting instead)")
			}
		}
	}

	return nil
}

// FrameBatch aplica framing a múltiples mensajes de forma eficiente
// Útil para procesar lotes de mensajes minimizando llamadas de sistema
//
// Esta función es más eficiente que llamar Frame() múltiples veces porque:
//   - Pre-calcula el tamaño del buffer necesario
//   - Reduce realocaciones de memoria
//   - Minimiza llamadas de sistema para escritura
func (f *Framer) FrameBatch(messages []string) ([]byte, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages to frame")
	}

	// Pre-calcular tamaño aproximado del buffer
	// Esto reduce significativamente las realocaciones de memoria
	estimatedSize := 0
	for _, msg := range messages {
		// Tamaño del mensaje + overhead del framing (estimado en 20 bytes)
		// El overhead incluye: longitud como string, espacio, y posible LF
		estimatedSize += len(msg) + 20
	}

	buf := bytes.NewBuffer(make([]byte, 0, estimatedSize))

	// Aplicar framing a cada mensaje
	for i, msg := range messages {
		framedMsg, err := f.Frame(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to frame message %d: %w", i, err)
		}
		buf.Write(framedMsg)
	}

	return buf.Bytes(), nil
}

// IsValidFramingMethod verifica si el método de framing es válido según RFC 6587
func IsValidFramingMethod(method FramingMethod) bool {
	return method == OctetCounting || method == NonTransparent
}

// RecommendedFramingMethod retorna el método de framing recomendado
// RFC 6587 recomienda octet-counting porque:
//   - No tiene restricciones sobre el contenido del mensaje
//   - Es más eficiente de parsear
//   - No requiere escaneo del contenido
func RecommendedFramingMethod() FramingMethod {
	return OctetCounting
}
