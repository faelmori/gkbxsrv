package utils

import (
	"context"
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/utils"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func CheckPortAsync(ctx context.Context, port string, listenerConfig net.ListenConfig) (<-chan string, <-chan error) {
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	go func() {
		listener, err := listenerConfig.Listen(ctx, "tcp", net.JoinHostPort("localhost", port))
		if err != nil {
			errorChan <- err
			return
		}
		defer func(listener net.Listener) {
			_ = listener.Close()
		}(listener)
		resultChan <- port
	}()
	return resultChan, errorChan
}

func IsPortAvailable(port string) (string, error) {
	listenerConfig := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			if err = c.Control(func(fd uintptr) {
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			}); err != nil {
				return err
			}
			return nil
		},
	}

	timeout := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan, errorChan := CheckPortAsync(ctx, port, listenerConfig)
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context deadline exceeded for port %s", port)
		case result := <-resultChan:
			return result, nil
		case err := <-errorChan:
			return "", err
		}
	}
}

func SearchAvailablePort(startPort string) (string, error) {
	intPort, err := strconv.Atoi(startPort)
	if err != nil {
		return "", fmt.Errorf("invalid port number: %s, error: %v", startPort, err)
	}

	maxAttempts := 65535 - intPort
	if maxAttempts > 100 {
		maxAttempts = 100
	}

	for attempts := 0; attempts < maxAttempts; attempts++ {
		availablePort := strconv.Itoa(intPort)
		if available, availableErr := IsPortAvailable(availablePort); availableErr == nil {
			return available, nil
		}
		intPort++
	}

	return "", fmt.Errorf("no available port found in range starting from %s", startPort)
}

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
