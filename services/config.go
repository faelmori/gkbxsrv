package services

import (
	cfgs "github.com/faelmori/gkbxsrv/internal/services"
)

type ConfigService = cfgs.IConfigService

func NewConfigService(configPath, keyPath, certPath string) ConfigService {
	return cfgs.NewConfigSrv(configPath, keyPath, certPath)
}
