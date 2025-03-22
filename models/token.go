package models

import (
	c "github.com/faelmori/gkbxsrv/internal/clientjwt"
	i "github.com/faelmori/gkbxsrv/internal/models"
	s "github.com/faelmori/gkbxsrv/services"
	l "github.com/faelmori/logz"
)

type TokenRepo interface {
	i.TokenRepo
}
type TokenService interface {
	i.TokenService
}

func LoadTokenCfg(cfgService s.ConfigService, fsService s.FilesystemService, crtService s.CertService, dbService s.DatabaseService) (TokenService, int64, int64, error) {
	ts := c.NewTokenClient(cfgService, fsService, &crtService, dbService)
	tokenService, idExpirationSecs, refExpirationSecs, loadTokenCfgError := ts.LoadTokenCfg()
	if loadTokenCfgError != nil {
		l.Error(loadTokenCfgError.Error(), nil)
		return nil, 0, 0, loadTokenCfgError
	}
	return *tokenService, idExpirationSecs, refExpirationSecs, nil
}
