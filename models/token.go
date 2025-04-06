package models

import (
	c "github.com/faelmori/gkbxsrv/internal/clientjwt"
	i "github.com/faelmori/gkbxsrv/internal/models"
	s "github.com/faelmori/kbxutils/utils/helpers"
)

type TokenRepo interface {
	i.TokenRepo
}
type TokenService interface {
	i.TokenService
}

func LoadTokenCfg(cfgService s.IConfigService, fsService s.FileSystemService, crtService s.ICertService, dbService s.IDatabaseService) (TokenService, int64, int64, error) {
	tkClient := c.NewTokenClient(cfgService, fsService, crtService, dbService)
	return tkClient.LoadTokenCfg()
}
