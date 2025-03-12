package services

import fsys "github.com/faelmori/gkbxsrv/internal/services"

type Broker = fsys.BrokerImpl
type BrokerInfo = fsys.BrokerInfoLock
type BrokerManager = fsys.BrokerManager

func NewBrokerService(verbose bool, port string) (*Broker, error) { return fsys.NewBroker(verbose) }
func NewBrokerManager() *BrokerManager                            { return fsys.NewBrokerManager() }
func NewBrokerInfo(port string) *BrokerInfo                       { return fsys.NewBrokerInfo("", port) }
