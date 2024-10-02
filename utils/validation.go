package utils

import (
	"net"
)

func IsValidAddress(address string) bool {
	_, _, err := net.SplitHostPort(address)
	return err == nil
}
