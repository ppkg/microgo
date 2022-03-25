package utils

import (
	"net"
	"strings"
)

func GetIp() string {
	ip := "127.0.0.1"
	conn, err := net.Dial("udp", "114.114.114.114:53")
	if err != nil {
		interfaces, _ := net.Interfaces()
		for _, inter := range interfaces {
			if addrs, err := inter.Addrs(); err == nil {
				for _, addr := range addrs {
					if addr.(*net.IPNet).IP.To4() != nil && addr.(*net.IPNet).IP.String() != "127.0.0.1" {
						if len(inter.Name) >= 1 && string(inter.Name[0]) == "e" {
							ip = addr.(*net.IPNet).IP.String()
							break
						}
					}
				}
			}
		}
	}
	defer conn.Close()
	ip = strings.Split(conn.LocalAddr().String(), ":")[0]
	return ip
}
