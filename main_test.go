package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yllada/Send-Log-TCP/routes"
)

func TestMain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.Default()

	routes.SetupRoutes(r)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), "<html")
}
