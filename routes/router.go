package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yllada/Send-Log-TCP/processor"
	"github.com/yllada/Send-Log-TCP/schema"
	"github.com/yllada/Send-Log-TCP/utils"
)

// SetupRoutes configura las rutas para el servidor Gin
func SetupRoutes(r *gin.Engine) {
	// Rutas para enviar logs
	r.POST("/sendlog", handleSendLog)
}

// handleSendLog maneja la recepción y procesamiento de logs
func handleSendLog(c *gin.Context) {
	var req schema.SyslogConfig

	// Vincular los datos de la solicitud a la estructura
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validar la dirección
	if !utils.IsValidAddressAndPort(req.Address, req.Port) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address or port format"})
		return
	}

	// Validar protocolo
	if req.Protocol != "tcp" && req.Protocol != "udp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Protocol must be tcp or udp"})
		return
	}

	// Establecer un intervalo por defecto si no se proporciona
	if req.Interval == 0 {
		req.Interval = 5 * time.Second
	}

	// Llamar a la función de envío de syslog
	go processor.SendSyslogMessages(req)

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{"status": "Logs sent"})
}
