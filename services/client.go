package services

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

const brokerSocketPath = "/tmp/GoLifeBroker.sck"

type IBrokerClient interface {
	SendCommand(command string) (string, error)
	Close() error
}

type BrokerClientImpl struct {
	Conn net.Conn
}

func (c *BrokerClientImpl) SendCommand(command string) (string, error) {
	_, err := c.Conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("erro ao enviar comando: %w", err)
	}

	reader := bufio.NewReader(c.Conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %w", err)
	}

	return strings.TrimSpace(response), nil
}
func (c *BrokerClientImpl) Close() error {
	return c.Conn.Close()
}

func NewBrokerClient() (IBrokerClient, error) {
	conn, err := net.Dial("unix", brokerSocketPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao broker: %w", err)
	}
	return &BrokerClientImpl{Conn: conn}, nil
}
