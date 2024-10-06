package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yllada/Send-Log-TCP/routes"
	"github.com/yllada/Send-Log-TCP/utils"
)

func main() {
	// Abrir el navegador en la URL http://localhost:8080
	utils.OpenBrowser("http://localhost:8080")

	time.Sleep(1 * time.Second)

	// Crear el router Gin
	r := gin.Default()

	// Configurar CORS
	r.Use(cors.Default())

	// Configurar rutas del API
	routes.SetupRoutes(r)

	// Iniciar servidor en el puerto 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
