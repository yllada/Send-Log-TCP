package utils

import (
	"net"
)

// isValidAddress valida que la dirección tenga el formato correcto
func IsValidAddress(address string) bool {
	_, _, err := net.SplitHostPort(address)
	return err == nil
}
