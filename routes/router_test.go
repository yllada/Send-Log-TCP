package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yllada/Send-Log-TCP/schema"
)

func mockIsValidAddressAndPort(address string, port string) bool {
	if address == "localhost" && port == "1234" {
		return true
	}
	return false
}

func TestHandleSendLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/sendlog", func(c *gin.Context) {
		handleSendLog(c, mockIsValidAddressAndPort)
	})

	tests := []struct {
		name         string
		reqBody      interface{} 
		expectedCode int
		expectedBody map[string]interface{}
	}{
		{
			name:         "Invalid JSON",
			reqBody:      "not a json", 
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"error": "Invalid request"},
		},
		{
			name: "Invalid Address and Port",
			reqBody: schema.SyslogConfig{
				Address:  "invalid_address",
				Port:     "not_a_port",
				Protocol: "tcp",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"error": "Invalid address or port format"},
		},
		{
			name: "Invalid Protocol",
			reqBody: schema.SyslogConfig{
				Address:  "localhost",
				Port:     "1234",
				Protocol: "http",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"error": "Protocol must be tcp or udp"},
		},
		{
			name: "Default Interval",
			reqBody: schema.SyslogConfig{
				Address:  "localhost",
				Port:     "1234",
				Protocol: "tcp",
				Interval: 0,
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{"status": "Logs sent"},
		},
		{
			name: "Successful Send Log",
			reqBody: schema.SyslogConfig{
				Address:  "localhost",
				Port:     "1234",
				Protocol: "tcp",
				Interval: 5 * time.Second,
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]interface{}{"status": "Logs sent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody) 
			req, _ := http.NewRequest(http.MethodPost, "/sendlog", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			var responseBody map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.Equal(t, tt.expectedBody, responseBody)
		})
	}
}
