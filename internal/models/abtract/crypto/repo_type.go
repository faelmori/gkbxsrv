package crypto

import (
	"context"
	"crypto/rsa"
	"gorm.io/gorm"
	"time"
)

type TSConfig struct {
	TokenRepository       RepoToken
	PrivKey               *rsa.PrivateKey
	PubKey                *rsa.PublicKey
	RefreshSecret         string
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}
type TokenPair struct {
	IDToken
	RefreshToken
}
type RefreshToken struct {
	ID  string `json:"-"`
	UID string `json:"-"`
	SS  string
}
type IDToken struct {
	SS string
}

type TypeRepoToken struct{ *gorm.DB }

func NewTokenRepo(db *gorm.DB) RepoToken { return &TypeRepoToken{db} }

func (g *TypeRepoToken) SetRefreshToken(ctx context.Context, userID string, tokenID string, expiresIn time.Duration) error {
	return nil
}
func (g *TypeRepoToken) DeleteRefreshToken(ctx context.Context, userID string, prevTokenID string) error {
	return nil
}
func (g *TypeRepoToken) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	return nil
}
