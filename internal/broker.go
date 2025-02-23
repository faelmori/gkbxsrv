package internal

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

const brokerSocketPath = "/tmp/GoLifeBroker.sck"

type BrokerClient interface {
	SendCommand(command string) (string, error)
	Close() error
}
type brokerClient struct {
	conn net.Conn
}

func (c *brokerClient) SendCommand(command string) (string, error) {
	_, err := c.conn.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("erro ao enviar comando: %w", err)
	}

	reader := bufio.NewReader(c.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %w", err)
	}

	return strings.TrimSpace(response), nil
}
func (c *brokerClient) Close() error {
	return c.conn.Close()
}

func NewBrokerClient() (BrokerClient, error) {
	conn, err := net.Dial("unix", brokerSocketPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao broker: %w", err)
	}
	return &brokerClient{conn: conn}, nil
}
