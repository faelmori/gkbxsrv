package models

import (
	"github.com/faelmori/gkbxsrv/internal/clientjwt"
	imodels "github.com/faelmori/gkbxsrv/internal/models"
	"github.com/faelmori/gkbxsrv/logz"
	"gorm.io/gorm"
)

type TSConfig = imodels.TSConfig
type TokenRepo struct{ imodels.TokenRepo }
type TokenService struct{ imodels.TokenService }

func LoadTokenCfg(db *gorm.DB) (*imodels.TokenService, int64, int64, error) {
	tc := clientjwt.NewTokenClient()
	tokenService, idExpirationSecs, refExpirationSecs, loadTokenCfgError := tc.LoadTokenCfg(db)
	if loadTokenCfgError != nil {
		logz.Logger.Error(loadTokenCfgError.Error(), nil)
		return nil, 0, 0, loadTokenCfgError
	}
	return tokenService, idExpirationSecs, refExpirationSecs, nil
}
