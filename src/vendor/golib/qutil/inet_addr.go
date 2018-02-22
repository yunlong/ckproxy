package qutil

import (
	"net"
	"strings"
)

func CIDR2IpRange(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	inc := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	ipRange := []string{}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ipRange = append(ipRange, ip.String())
	}

	return ipRange, nil
}

func GetInternalNetworkIP() string {
	info, _ := net.InterfaceAddrs()
	for _, addr := range info {
		ip := strings.Split(addr.String(), "/")[0]
		if strings.HasPrefix(ip, "10.") {
			return ip
		}
	}

	return ""
}
