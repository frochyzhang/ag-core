package ip

import (
	"fmt"
	"net"
	"strings"
)

const (
	minPort int = 0
	maxPort int = 65535
)

// GetAvailablePort 检查当前指定端口是否可用，不可用则自动+1再试（随机端口从默认端口开始检查）
func GetAvailablePort(host string, oriPort int) (int, error) {

	// 检查host是否可用的
	if !IsHostAvailable(host) {
		return 0, fmt.Errorf("ERROR_HOST_NOT_FOUND:%s", host)
	}

	port := oriPort
	if port < minPort {
		port = minPort
	}

	for port < maxPort {
		if isPortAvailable(host, port) {
			return port, nil
		}
		port++
	}

	return 0, fmt.Errorf("ERROR_BIND_PORT_ERROR")
}

// isPortAvailable 检查指定主机上的给定端口是否可用
func isPortAvailable(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

// IsHostAvailable 检查host是否可用
func IsHostAvailable(host string) bool {
	// isAnyHost
	if strings.EqualFold("0.0.0.0", host) {
		return true
	}
	// isLocalHost
	if strings.EqualFold("127.0.0.1", host) || strings.EqualFold("localhost", host) {
		return true
	}

	// isHostInNetWork 检查
	return isHostInNetworkCard(host)
}

// isHostInNetworkCard 是否网卡上的地址
func isHostInNetworkCard(host string) bool {
	addr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return false
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return false
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				if ipnet.IP.Equal(addr.IP) {
					return true
				}
			}
		}
	}

	return false
}

func ParseTcpIPAndPort(input string) (string, int, error) {
	parts := strings.Split(input, ":")
	ip := parts[0]
	port := 0

	var err error

	if len(parts) > 1 {
		portStr := parts[1]
		port, err = net.LookupPort("tcp", portStr)
	}

	if net.ParseIP(ip) == nil {
		ip = ""
		// err = fmt.Errorf("Invalid IP address: %s", ip)
	}

	return ip, port, err
}
