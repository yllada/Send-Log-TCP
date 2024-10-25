package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// App struct
type App struct {
	ctx         context.Context
	conn        net.Conn
	connMu      sync.Mutex // Mutex para manejar el acceso a la conexión
	isConnected bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {}

// beforeClose is called when the application is about to quit.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {}

type SyslogConfig struct {
	Address  string
	Port     string
	Protocol string
	Messages []string
}

// SyslogResponse to send back to the frontend
type SyslogResponse struct {
	SentMessages []string `json:"sentMessages"`
	Errors       []string `json:"errors"`
}

// CheckConnection verifies if a connection can be established
func (a *App) CheckConnection(address string, port string, protocol string) (bool, error) {
	fullAddress := fmt.Sprintf("%s:%s", address, port)
	conn, err := net.Dial(protocol, fullAddress)
	if err != nil {
		log.Printf("Error connecting to %s: %v", fullAddress, err)
		return false, err
	}
	// Almacenar la conexión activa
	a.connMu.Lock()
	a.conn = conn
	a.isConnected = true
	a.connMu.Unlock()

	log.Println("Connected to", fullAddress)
	return true, nil
}

// Disconnect cierra la conexión activa si existe
func (a *App) Disconnect() {
	a.connMu.Lock()
	defer a.connMu.Unlock()

	if a.isConnected {
		err := a.conn.Close()
		if err != nil {
			log.Printf("Error disconnecting: %v", err)
		} else {
			log.Println("Disconnected successfully.")
		}
		a.isConnected = false
	}
}

// SendSyslogMessages sends syslog messages and returns the result
func (a *App) SendSyslogMessages(config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	fullAddress := fmt.Sprintf("%s:%s", config.Address, config.Port)

	conn, err := net.Dial(config.Protocol, fullAddress)
	if err != nil {
		log.Printf("Error connecting to %s: %v", fullAddress, err)
		response.Errors = append(response.Errors, fmt.Sprintf("Error connecting to %s: %v", fullAddress, err))
		return response
	}
	log.Println("Connected to", fullAddress)
	defer conn.Close()

	for _, message := range config.Messages {
		syslogMessage := BuildSyslogMessage(config, message)
		_, err := conn.Write([]byte(syslogMessage))
		if err != nil {
			log.Printf("Error sending message: %v", err)
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			log.Printf("Sent message: %s", syslogMessage)
			response.SentMessages = append(response.SentMessages, syslogMessage)
		}
	}

	return response
}

func BuildSyslogMessage(config SyslogConfig, message string) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")
	return fmt.Sprintf("%d %s %s\r\n", 679, timestamp, message)
}
