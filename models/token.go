package models

import (
	c "github.com/faelmori/gkbxsrv/internal/clientjwt"
	i "github.com/faelmori/gkbxsrv/internal/models"
	"github.com/faelmori/kbxutils/utils/interfaces"
)

type TokenRepo interface {
	i.TokenRepo
}
type TokenService interface {
	i.TokenService
}

func LoadTokenCfg(cfgService interfaces.IConfigService, fsService interfaces.FileSystemService, crtService interfaces.ICertService, dbService interfaces.IDatabaseService) (TokenService, int64, int64, error) {
	tkClient := c.NewTokenClient(cfgService, fsService, crtService, dbService)
	return tkClient.LoadTokenCfg()
}
