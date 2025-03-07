package crypto

import (
	"context"

	m "github.com/faelmori/gkbxsrv/internal/models/abtract/users"
)

type TokenService interface {
	NewPairFromUser(ctx context.Context, u m.User, prevTokenID string) (*TokenPair, error)
	SignOut(ctx context.Context, uid string) error
	ValidateIDToken(tokenString string) (m.User, error)
	ValidateRefreshToken(refreshTokenString string) (*RefreshToken, error)
	RenewToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}
