package utils

import (
	"fmt"
	"net"
)

// IsValidAddressAndPort valida que la dirección y puerto tengan el formato correcto
func IsValidAddressAndPort(address string, port string) bool {
	// Concatenar address y port
	fullAddress := fmt.Sprintf("%s:%s", address, port)

	// Validar el formato
	_, _, err := net.SplitHostPort(fullAddress)
	return err == nil
}
