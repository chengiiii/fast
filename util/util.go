package util

import (
	"net"
)

func GetOutboundIp() net.IP {
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return net.IPv4(127, 0, 0, 1)
}
