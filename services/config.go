package services

import (
	"fmt"
	gbls "github.com/faelmori/gkbxsrv/internal/globals"
	cfgs "github.com/faelmori/gkbxsrv/internal/services"
	l "github.com/faelmori/logz"
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

func GetServerConfig(configService ConfigService) *gbls.Server {
	serverConfig, serverConfigErr := configService.GetSettings()
	if serverConfigErr != nil {
		l.Error("Erro ao obter as configurações do servidor", map[string]interface{}{"error": serverConfigErr.Error()})
		return nil
	}
	for k, v := range serverConfig {
		if k == "server" {
			serverConfig = v.(map[string]interface{})
		}
	}
	if serverConfig != nil {
		srv := &gbls.Server{}
		srv.Port = serverConfig["port"].(string)
		srv.BindAddress = serverConfig["bind_address"].(string)
		srv.ReadTimeout = serverConfig["read_timeout"].(int)
		srv.WriteTimeout = serverConfig["write_timeout"].(int)
		return srv
	}
	return nil
}
