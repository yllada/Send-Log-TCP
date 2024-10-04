package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yllada/Send-Log-TCP/routes"
)

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode) // Set Gin to Test Mode
	r := gin.Default()
	routes.SetupRoutes(r)

	req, err := http.NewRequest(http.MethodPost, "/sendlog", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 but got %d\n", w.Code)
	}
}
