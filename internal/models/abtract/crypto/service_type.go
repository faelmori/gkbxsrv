package crypto

import (
	"context"
	"crypto/rsa"
	"fmt"
	m "github.com/faelmori/gkbxsrv/internal/models/abtract/users"
	"github.com/google/uuid"
	"strings"
)

func NewTokenService(c *TSConfig) TokenService {
	return &TypeTokenService{
		TokenRepository:       c.TokenRepository,
		PrivKey:               c.PrivKey,
		PubKey:                c.PubKey,
		RefreshSecret:         c.RefreshSecret,
		IDExpirationSecs:      c.IDExpirationSecs,
		RefreshExpirationSecs: c.RefreshExpirationSecs,
	}
}

type TypeTokenService struct {
	TokenRepository       RepoToken
	PrivKey               *rsa.PrivateKey
	PubKey                *rsa.PublicKey
	RefreshSecret         string
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

func (s *TypeTokenService) NewPairFromUser(ctx context.Context, u m.User, prevTokenID string) (*TokenPair, error) {
	if prevTokenID != "" {
		if err := s.TokenRepository.DeleteRefreshToken(ctx, u.GetID(), prevTokenID); err != nil {
			//return nil, logz.ErrorLog(fmt.Sprintf("Could not delete previous refreshToken for uid: %v, tokenID: %v\n", u.GetID(), prevTokenID), "GoSpyder")
			return nil, fmt.Errorf("Could not delete previous refreshToken for uid: %v, tokenID: %v\n", u.GetID(), prevTokenID)
		}
	}

	idToken, err := generateIDToken(u, s.PrivKey, s.IDExpirationSecs)
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error generating idToken for uid: %v. Error: %v\n", u.GetID(), err.Error()), "GoSpyder")
		return nil, fmt.Errorf("Error generating idToken for uid: %v. Error: %v\n", u.GetID(), err.Error())
	}

	refreshToken, err := generateRefreshToken(u.GetID(), s.RefreshSecret, s.RefreshExpirationSecs)
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error generating refreshToken for uid: %v. Error: %v\n", u.GetID(), err.Error()), "GoSpyder")
		return nil, fmt.Errorf("Error generating refreshToken for uid: %v. Error: %v\n", u.GetID(), err.Error())
	}

	if err := s.TokenRepository.SetRefreshToken(ctx, u.GetID(), refreshToken.ID, refreshToken.ExpiresIn); err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error storing tokenID for uid: %v. Error: %v\n", u.GetID(), err), "GoSpyder")
		return nil, fmt.Errorf("Error storing tokenID for uid: %v. Error: %v\n", u.GetID(), err)
	}

	return &TokenPair{
		IDToken:      IDToken{SS: idToken},
		RefreshToken: RefreshToken{SS: refreshToken.SS, ID: refreshToken.ID, UID: u.GetID()},
	}, nil
}
func (s *TypeTokenService) SignOut(ctx context.Context, uid string) error {
	return s.TokenRepository.DeleteUserRefreshTokens(ctx, uid)
}
func (s *TypeTokenService) ValidateIDToken(tokenString string) (m.User, error) {
	claims, err := validateIDToken(tokenString, s.PubKey)
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Unable to validate or parse idToken - Error: %v\n", err), "GoSpyder")
		return nil, fmt.Errorf("Unable to validate or parse idToken - Error: %v\n", err)
	}
	return claims.User, nil
}
func (s *TypeTokenService) ValidateRefreshToken(tokenString string) (*RefreshToken, error) {
	claims, claimsErr := validateRefreshToken(tokenString, s.RefreshSecret)
	if claimsErr != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Unable to validate or parse refreshToken for token string: %s\n%v\n", tokenString, claimsErr), "GoSpyder")
		return nil, fmt.Errorf("Unable to validate or parse refreshToken for token string: %s\n%v\n", tokenString, claimsErr)
	}
	tokenUUID, tokenUUIDErr := uuid.Parse(claims.Id)
	if tokenUUIDErr != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Claims ID could not be parsed as UUID: %s\n%v\n", claims.UID, tokenUUIDErr), "GoSpyder")
		return nil, fmt.Errorf("Claims ID could not be parsed as UUID: %s\n%v\n", claims.UID, tokenUUIDErr)
	}
	return &RefreshToken{
		SS:  tokenString,
		ID:  tokenUUID.String(),
		UID: claims.UID,
	}, nil
}
func (s *TypeTokenService) RenewToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if len(strings.Split(refreshToken, ".")) != 3 {
		//return nil, logz.ErrorLog(fmt.Sprintf("Invalid refreshToken format for token string: %s\n", refreshToken), "GoSpyder")
		return nil, fmt.Errorf("Invalid refreshToken format for token string: %s\n", refreshToken)
	}

	claims, err := validateRefreshToken(refreshToken, s.RefreshSecret)
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Unable to validate or parse refreshToken for token string: %s\n%v\n", refreshToken, err), "GoSpyder")
		return nil, fmt.Errorf("Unable to validate or parse refreshToken for token string: %s\n%v\n", refreshToken, err)
	}
	if err := s.TokenRepository.DeleteRefreshToken(ctx, claims.UID, claims.Id); err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error deleting refresh token: %v\n", err), "GoSpyder")
		return nil, fmt.Errorf("Error deleting refresh token: %v\n", err)
	}
	idCClaims, idCClaimsErr := validateIDToken(claims.UID, s.PubKey)
	if idCClaimsErr != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("Error validating idToken: %v\n", idCClaimsErr), "GoSpyder")
		return nil, fmt.Errorf("Error validating idToken: %v\n", idCClaimsErr)
	}
	return s.NewPairFromUser(ctx, idCClaims.User, claims.Id)
}
