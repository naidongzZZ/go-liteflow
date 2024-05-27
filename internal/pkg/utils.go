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

    ips, err := net.LookupIP(host)
    if err != nil || len(ips) == 0 {
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