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
	// Configurar Gin en modo de prueba
	gin.SetMode(gin.TestMode)

	// Crear un nuevo router
	r := gin.Default()

	// Configurar las rutas
	routes.SetupRoutes(r)

	// Crear un recorder para capturar la respuesta
	w := httptest.NewRecorder()

	// Realizar una solicitud GET a la ruta que sirve el index.html
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	// Verificar que el código de estado sea 200 OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Verificar que el contenido de la respuesta sea el esperado
	assert.Contains(t, w.Body.String(), "<html") // Verifica que contenga una etiqueta HTML
}

func TestAssets(t *testing.T) {
	// Configurar Gin en modo de prueba
	gin.SetMode(gin.TestMode)

	// Crear un nuevo router
	r := gin.Default()

	// Configurar las rutas
	routes.SetupRoutes(r)

	// Crear un recorder para capturar la respuesta
	w := httptest.NewRecorder()

	// Realizar una solicitud GET a la ruta de assets
	req, _ := http.NewRequest(http.MethodGet, "/assets/", nil)
	r.ServeHTTP(w, req)

	// Verificar que el código de estado sea 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
}
