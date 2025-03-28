package models

import (
	c "github.com/faelmori/gkbxsrv/internal/clientjwt"
	i "github.com/faelmori/gkbxsrv/internal/models"
	s "github.com/faelmori/gkbxsrv/services"
)

type TokenRepo interface {
	i.TokenRepo
}
type TokenService interface {
	i.TokenService
}

func LoadTokenCfg(cfgService s.ConfigService, fsService s.FilesystemService, crtService s.CertService, dbService s.DatabaseService) (TokenService, int64, int64, error) {
	tkClient := c.NewTokenClient(cfgService, fsService, crtService, dbService)
	return tkClient.LoadTokenCfg()
}
