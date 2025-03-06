package services

import fsys "github.com/faelmori/gkbxsrv/internal/services"

type Broker = fsys.BrokerImpl

func NewBrokerService(verbose bool) (*Broker, error) {
	return fsys.NewBroker(verbose)
}
