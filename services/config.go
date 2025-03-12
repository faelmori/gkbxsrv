package services

import (
	"fmt"
	gbls "github.com/faelmori/gkbxsrv/internal/globals"
	cfgs "github.com/faelmori/gkbxsrv/internal/services"
)

type Database = gbls.Database
type ConfigService = cfgs.ConfigService

func NewConfigService(configPath, keyPath, certPath string) ConfigService {
	return cfgs.NewConfigSrv(configPath, keyPath, certPath)
}

func GetDatabaseConfig(configService ConfigService) *Database {
	dbConfig := configService.GetDatabaseConfig()
	return &dbConfig
}

func SetDatabaseConfig(configService ConfigService, dbConfig *Database) error {
	if dbConfig.Port.(int) <= 0 || dbConfig.Port.(int) > 65535 {
		return fmt.Errorf("porta inválida: %v", dbConfig.Port)
	}
	if dbConfig.Host == "" {
		return fmt.Errorf("host não pode estar vazio")
	}
	return configService.SetDatabaseConfig(dbConfig)
}
