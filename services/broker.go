package services

import fsys "github.com/faelmori/gkbxsrv/internal/services"

type Broker interface {
	fsys.Broker
}

func NewBrokerService(cfgSrv ConfigService) Broker {
	return fsys.NewBroker(cfgSrv)
}
