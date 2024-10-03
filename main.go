package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/yllada/Send-Log-TCP/routes"
)

func main() {
	// Crear el router Gin
	r := gin.Default()

	// Configurar rutas del API
	routes.SetupRoutes(r)

	// Iniciar servidor en el puerto 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
