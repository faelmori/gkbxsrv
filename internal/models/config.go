package models

import (
	"context"
	"crypto/rsa"
	"gorm.io/gorm"
	"time"
)

type TokenRepo interface {
	SetRefreshToken(ctx context.Context, userID string, tokenID string, expiresIn time.Duration) error
	DeleteRefreshToken(ctx context.Context, userID string, prevTokenID string) error
	DeleteUserRefreshTokens(ctx context.Context, userID string) error
}

type TSConfig struct {
	TokenRepository       TokenRepo
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

type TokenRepoImpl struct{ *gorm.DB }

func NewTokenRepo(db *gorm.DB) TokenRepo { return &TokenRepoImpl{db} }

func (g *TokenRepoImpl) SetRefreshToken(ctx context.Context, userID string, tokenID string, expiresIn time.Duration) error {
	return nil
}
func (g *TokenRepoImpl) DeleteRefreshToken(ctx context.Context, userID string, prevTokenID string) error {
	return nil
}
func (g *TokenRepoImpl) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	return nil
}
