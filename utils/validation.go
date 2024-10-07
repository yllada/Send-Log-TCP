package utils

import (
	"fmt"
	"net"
)

func IsValidAddressAndPort(address string, port string) bool {
	fullAddress := fmt.Sprintf("%s:%s", address, port)

	_, _, err := net.SplitHostPort(fullAddress)
	return err == nil
}
