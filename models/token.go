package models

import (
	imodels "github.com/faelmori/gokubexfs/internal/models"
	"gorm.io/gorm"
)

type TSConfig = imodels.TSConfig
type TokenRepo struct{ imodels.TokenRepo }
type TokenService struct{ imodels.TokenService }

func NewTokenRepo(db *gorm.DB) *TokenRepo { return &TokenRepo{imodels.NewTokenRepo(db)} }
func NewTokenService(config *imodels.TSConfig) *TokenService {
	return &TokenService{imodels.NewTokenService(config)}
}
func LoadTokenCfg(db *gorm.DB) (imodels.TokenService, int64, int64, error) {
	nt := NewTokenRepo(db)
	ts := NewTokenService(&imodels.TSConfig{TokenRepository: nt})
	idExpirationSecs := int64(0)
	refExpirationSecs := int64(0)

	return ts, idExpirationSecs, refExpirationSecs, nil
}
