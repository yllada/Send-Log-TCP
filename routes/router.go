package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yllada/Send-Log-TCP/processor"
	"github.com/yllada/Send-Log-TCP/schema"
	"github.com/yllada/Send-Log-TCP/utils"
)

func SetupRoutes(r *gin.Engine) {
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")

	r.POST("/sendlog", func(c *gin.Context) {
		handleSendLog(c, utils.IsValidAddressAndPort)
	})
}

func handleSendLog(c *gin.Context, isValidAddressAndPort func(string, string) bool) {
	var req schema.SyslogConfig

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if !isValidAddressAndPort(req.Address, req.Port) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address or port format"})
		return
	}

	if req.Protocol != "tcp" && req.Protocol != "udp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Protocol must be tcp or udp"})
		return
	}

	if req.Interval == 0 {
		req.Interval = 5 * time.Second
	}

	go processor.SendSyslogMessages(req)

	c.JSON(http.StatusOK, gin.H{"status": "Logs sent"})
}
