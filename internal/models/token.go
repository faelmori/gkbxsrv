package models

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	//"github.com/faelmori/logz"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

type idTokenCustomClaims struct {
	User User `json:"UserImpl"`
	jwt.StandardClaims
}
type TokenService interface {
	NewPairFromUser(ctx context.Context, u User, prevTokenID string) (*TokenPair, error)
	SignOut(ctx context.Context, uid string) error
	ValidateIDToken(tokenString string) (User, error)
	ValidateRefreshToken(refreshTokenString string) (*RefreshToken, error)
	RenewToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}
type TokenServiceImpl struct {
	TokenRepository       TokenRepo
	PrivKey               *rsa.PrivateKey
	PubKey                *rsa.PublicKey
	RefreshSecret         string
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

func NewTokenService(c *TSConfig) TokenService {
	return &TokenServiceImpl{
		TokenRepository:       c.TokenRepository,
		PrivKey:               c.PrivKey,
		PubKey:                c.PubKey,
		RefreshSecret:         c.RefreshSecret,
		IDExpirationSecs:      c.IDExpirationSecs,
		RefreshExpirationSecs: c.RefreshExpirationSecs,
	}
}

func (s *TokenServiceImpl) NewPairFromUser(ctx context.Context, u User, prevTokenID string) (*TokenPair, error) {
	if prevTokenID != "" {
		if err := s.TokenRepository.DeleteRefreshToken(ctx, u.GetID(), prevTokenID); err != nil {
			//return nil, logz.ErrorLog(fmt.Sprintf("Could not delete previous refreshToken for uid: %v, tokenID: %v\n", u.GetID(), prevTokenID), "GoSpider")
			return nil, fmt.Errorf("Could not delete previous refreshToken for uid: %v, tokenID: %v\n", u.GetID(), prevTokenID)
		}
	}

	idToken, err := generateIDToken(u, s.PrivKey, s.IDExpirationSecs)
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error generating idToken for uid: %v. Error: %v\n", u.GetID(), err.Error()), "GoSpider")
		return nil, fmt.Errorf("Error generating idToken for uid: %v. Error: %v\n", u.GetID(), err.Error())
	}

	refreshToken, err := generateRefreshToken(u.GetID(), s.RefreshSecret, s.RefreshExpirationSecs)
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error generating refreshToken for uid: %v. Error: %v\n", u.GetID(), err.Error()), "GoSpider")
		return nil, fmt.Errorf("Error generating refreshToken for uid: %v. Error: %v\n", u.GetID(), err.Error())
	}

	if err := s.TokenRepository.SetRefreshToken(ctx, u.GetID(), refreshToken.ID, refreshToken.ExpiresIn); err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error storing tokenID for uid: %v. Error: %v\n", u.GetID(), err), "GoSpider")
		return nil, fmt.Errorf("Error storing tokenID for uid: %v. Error: %v\n", u.GetID(), err)
	}

	return &TokenPair{
		IDToken:      IDToken{SS: idToken},
		RefreshToken: RefreshToken{SS: refreshToken.SS, ID: refreshToken.ID, UID: u.GetID()},
	}, nil
}
func (s *TokenServiceImpl) SignOut(ctx context.Context, uid string) error {
	return s.TokenRepository.DeleteUserRefreshTokens(ctx, uid)
}
func (s *TokenServiceImpl) ValidateIDToken(tokenString string) (User, error) {
	claims, err := validateIDToken(tokenString, s.PubKey)
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Unable to validate or parse idToken - Error: %v\n", err), "GoSpider")
		return nil, fmt.Errorf("Unable to validate or parse idToken - Error: %v\n", err)
	}
	return claims.User, nil
}
func (s *TokenServiceImpl) ValidateRefreshToken(tokenString string) (*RefreshToken, error) {
	claims, claimsErr := validateRefreshToken(tokenString, s.RefreshSecret)
	if claimsErr != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Unable to validate or parse refreshToken for token string: %s\n%v\n", tokenString, claimsErr), "GoSpider")
		return nil, fmt.Errorf("Unable to validate or parse refreshToken for token string: %s\n%v\n", tokenString, claimsErr)
	}
	tokenUUID, tokenUUIDErr := uuid.Parse(claims.Id)
	if tokenUUIDErr != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Claims ID could not be parsed as UUID: %s\n%v\n", claims.UID, tokenUUIDErr), "GoSpider")
		return nil, fmt.Errorf("Claims ID could not be parsed as UUID: %s\n%v\n", claims.UID, tokenUUIDErr)
	}
	return &RefreshToken{
		SS:  tokenString,
		ID:  tokenUUID.String(),
		UID: claims.UID,
	}, nil
}
func (s *TokenServiceImpl) RenewToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if len(strings.Split(refreshToken, ".")) != 3 {
		//return nil, logz.ErrorLog(fmt.Sprintf("Invalid refreshToken format for token string: %s\n", refreshToken), "GoSpider")
		return nil, fmt.Errorf("Invalid refreshToken format for token string: %s\n", refreshToken)
	}

	claims, err := validateRefreshToken(refreshToken, s.RefreshSecret)
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Unable to validate or parse refreshToken for token string: %s\n%v\n", refreshToken, err), "GoSpider")
		return nil, fmt.Errorf("Unable to validate or parse refreshToken for token string: %s\n%v\n", refreshToken, err)
	}
	if err := s.TokenRepository.DeleteRefreshToken(ctx, claims.UID, claims.Id); err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error deleting refresh token: %v\n", err), "GoSpider")
		return nil, fmt.Errorf("Error deleting refresh token: %v\n", err)
	}
	idCClaims, idCClaimsErr := validateIDToken(claims.UID, s.PubKey)
	if idCClaimsErr != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error validating idToken: %v\n", idCClaimsErr), "GoSpider")
		return nil, fmt.Errorf("Error validating idToken: %v\n", idCClaimsErr)
	}
	return s.NewPairFromUser(ctx, idCClaims.User, claims.Id)
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

func generateIDToken(u User, key *rsa.PrivateKey, exp int64) (string, error) {
	unixTime := time.Now().Unix()
	tokenExp := unixTime + exp
	claims := idTokenCustomClaims{
		User: u,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  unixTime,
			ExpiresAt: tokenExp,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		//return "", logz.ErrorLog(fmt.Sprintf("Failed to sign id token string"), "GoSpider")
		return "", fmt.Errorf("Failed to sign id token string")
	}

	return ss, nil
}
func generateRefreshToken(uid string, key string, exp int64) (*refreshTokenData, error) {
	currentTime := time.Now()
	tokenExp := currentTime.Add(time.Duration(exp) * time.Second)
	tokenID, err := uuid.NewRandom()
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Failed to generate refresh token ID"), "GoSpider")
		return nil, fmt.Errorf("Failed to generate refresh token ID")
	}

	claims := refreshTokenCustomClaims{
		UID: uid,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  currentTime.Unix(),
			ExpiresAt: tokenExp.Unix(),
			Id:        tokenID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(key))
	if err != nil {
		log.Println("Failed to sign refresh token string")
		return nil, err
	}

	return &refreshTokenData{
		SS:        ss,
		ID:        tokenID.String(),
		ExpiresIn: tokenExp.Sub(currentTime),
	}, nil
}
func validateIDToken(tokenString string, key *rsa.PublicKey) (*idTokenCustomClaims, error) {
	claims := &idTokenCustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("ID token is invalid")
	}
	claims, ok := token.Claims.(*idTokenCustomClaims)
	if !ok {
		return nil, fmt.Errorf("ID token valid but couldn't parse claims")
	}
	return claims, nil
}
func validateRefreshToken(tokenString string, key string) (*refreshTokenCustomClaims, error) {
	claims := &refreshTokenCustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("refresh token is invalid")
	}
	claims, ok := token.Claims.(*refreshTokenCustomClaims)
	if !ok {
		return nil, fmt.Errorf("refresh token valid but couldn't parse claims")
	}
	return claims, nil
}
