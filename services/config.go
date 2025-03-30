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
		prt := serverConfig["port"]
		if prt == nil {
			prt = "8888"
		}
		bda := serverConfig["bind_address"]
		if bda == nil {
			bda = "0.0.0.0"
		}
		rto := serverConfig["read_timeout"]
		if rto == nil {
			rto = 60
		}
		wto := serverConfig["write_timeout"]
		if wto == nil {
			wto = 60
		}

		srv := &gbls.Server{}
		srv.Port = prt.(string)
		srv.BindAddress = bda.(string)
		srv.ReadTimeout = int(rto.(float64))
		srv.WriteTimeout = int(wto.(float64))

		return srv
	}
	return nil
}
