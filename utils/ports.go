package utils

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/utils"
	"net"
	"os/exec"
	"strings"
)

func CheckPortOpen(port string) bool {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func ListOpenPorts() ([]string, error) {
	var openPorts []string
	for port := 1; port <= 65535; port++ {
		if CheckPortOpen(fmt.Sprintf("%d", port)) {
			openPorts = append(openPorts, fmt.Sprintf("%d", port))
		}
	}
	return openPorts, nil
}

func ClosePort(port string) error {
	cmd := exec.Command("fuser", "-k", port+"/tcp")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("falha ao fechar a porta %s: %v", port, err)
	}
	return nil
}

func OpenPort(port string) error {
	cmd := exec.Command("iptables", "-A", "INPUT", "-p", "tcp", "--dport", port, "-j", "ACCEPT")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("falha ao abrir a porta %s: %v", port, err)
	}
	return nil
}

func IsIPv6(ip string) bool {
	var sanitizedIP string
	if strings.Contains(ip, ":") {
		sanitizedIP = strings.Replace(ip, ":", "", -1)

		// If the sanitized IP is still the same as the original, then it's an IPv6
		if sanitizedIP == ip {
			return true
		}
	}
	return false
}

func findAvailablePort(basePort int, maxAttempts int) (string, error) {
	for i := 0; i < maxAttempts; i++ {
		port := fmt.Sprintf("%d", basePort+i)
		isOpen, _ := utils.CheckPortOpen(port) // Usando função do utils
		if !isOpen {
			return port, nil
		}
	}
	return "", fmt.Errorf("nenhuma porta disponível no range %d-%d", basePort, basePort+maxAttempts-1)
}
