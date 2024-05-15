package pkg

import (
	"net"
	"strconv"
)

func ValidateIPv4WithPort(address string) bool {
    host, port, err := net.SplitHostPort(address)
    if err != nil {
        return false
    }

    ip := net.ParseIP(host)
    if ip.To4() == nil {
        return false
    }

	portNum, err := strconv.Atoi(port)
    if err != nil {
        return false
    }
    if portNum < 1 || portNum > 65535 {
        return false
    }
    return true
}