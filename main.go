package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/yllada/send-log-tcp/config"
	"github.com/yllada/send-log-tcp/utils"
)

const maxRetries = 3
const retryDelay = 2 * time.Second

func main() {
	address, logMessage := config.GetConfig()

	if !utils.IsValidAddress(address) {
		log.Fatalf("Invalid IP address or port: %s", address)
	}

	var conn net.Conn
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = net.Dial("tcp", address)
		if err == nil {
			break
		}
		log.Printf("Error connecting (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		log.Fatalf("Could not connect after %d attempts: %v", maxRetries, err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(logMessage))
	if err != nil {
		log.Fatalf("Error al enviar el log: %v", err)
	}

	fmt.Println("Log send correctily")
}
