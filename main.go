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
	utils.OpenBrowser("http://localhost:8080")

	time.Sleep(1 * time.Second)

	r := gin.Default()

	r.Use(cors.Default())

	routes.SetupRoutes(r)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
