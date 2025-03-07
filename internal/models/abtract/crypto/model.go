package crypto

import (
	"github.com/dgrijalva/jwt-go"
	m "github.com/faelmori/gkbxsrv/internal/models/abtract/users"
	"time"
)

type idTokenCustomClaims struct {
	User m.User `json:"UserImpl"`
	jwt.StandardClaims
}
type refreshTokenData struct {
	SS        string
	ID        string
	ExpiresIn time.Duration
}
type refreshTokenCustomClaims struct {
	UID string `json:"uid"`
	jwt.StandardClaims
}
