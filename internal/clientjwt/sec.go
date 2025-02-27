package clientjwt

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	imodels "github.com/faelmori/gkbxsrv/internal/models"
	"github.com/faelmori/gkbxsrv/services"
	srvs "github.com/faelmori/gkbxsrv/services"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"log"
)

var cfg srvs.ConfigService
var fs srvs.FilesystemService
var crt *srvs.CertService

type TokenClient interface {
	LoadPrivateKey() (*rsa.PrivateKey, error)
	LoadPublicKey(cfg services.ConfigService) *rsa.PublicKey
	LoadTokenCfg(db *gorm.DB) (*imodels.TokenService, int64, int64, error)
}

type TokenClientImpl struct {
	cfg                   srvs.ConfigService
	fs                    srvs.FilesystemService
	crt                   *srvs.CertService
	TokenService          *imodels.TokenService
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

func (t *TokenClientImpl) LoadPrivateKey() (*rsa.PrivateKey, error) {
	keyDataB64 := viper.GetString("jwt.private_key")
	if keyDataB64 == "" {
		return nil, fmt.Errorf("error reading private key file: %v", "jwt.private_key is empty")
	}
	keyData, err := base64.StdEncoding.DecodeString(keyDataB64)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v", err)
	}
	crtB := *crt
	pwd, pwdErr := crtB.RetrievePassword()
	if pwdErr != nil {
		return nil, fmt.Errorf("error reading private key file: %v", pwdErr)
	}
	privateKey, privateKeyErr := crtB.DecryptPrivateKey(keyData, []byte(pwd))
	if privateKeyErr != nil {
		return nil, fmt.Errorf("error parsing private key: %v", privateKeyErr)
	}
	return privateKey, nil
}
func (t *TokenClientImpl) LoadPublicKey(cfg services.ConfigService) *rsa.PublicKey {
	if readErr := viper.ReadInConfig(); readErr != nil {
		log.Fatalf("Error reading public key file: %v", readErr)
	}
	if !fs.ExistsConfigFile() {
		log.Fatalf("Error reading public key file: %v", "config file not found")
	}

	if !cfg.IsConfigLoaded() {
		if loadErr := cfg.LoadConfig(); loadErr != nil {
			log.Fatalf("Error loading config file: %v", loadErr)
		}
	}

	keyDataB64 := viper.GetString("jwt.public_key")
	if keyDataB64 == "" {
		log.Fatalf("Error reading public key file: %v", "jwt.public_key is empty")
	}

	keyData, err := base64.StdEncoding.DecodeString(keyDataB64)
	if err != nil {
		log.Fatalf("Error reading public key file: %v", err)
	}

	publicKey, publicKeyErr := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if publicKeyErr != nil {
		fmt.Println(publicKey)
		fmt.Println(publicKeyErr)
		log.Fatalf("Error parsing public key: %v", publicKeyErr)
	}

	return publicKey
}
func (t *TokenClientImpl) LoadTokenCfg(db *gorm.DB) (*imodels.TokenService, int64, int64, error) {
	privKey := viper.GetString("jwt.private_key")
	if privKey == "" {
		return nil, 0, 0, fmt.Errorf("error reading private key file: %v", "jwt.private_key is empty")
	}
	rsaPrivKey, rsaPrivKeyErr := t.LoadPrivateKey()
	if rsaPrivKeyErr != nil {
		return nil, 0, 0, rsaPrivKeyErr
	}
	if validateErr := rsaPrivKey.Validate(); validateErr != nil {
		return nil, 0, 0, fmt.Errorf("error reading private key file: %v", validateErr)
	}
	pubKey := &rsa.PublicKey{
		N: rsaPrivKey.N,
		E: rsaPrivKey.E,
	}
	if !pubKey.Equal(rsaPrivKey.Public()) {
		return nil, 0, 0, fmt.Errorf("error reading private key file: %v", "public key does not match private key")
	}
	refreshSecret := viper.GetString("jwt.refresh_secret")
	idExpirationSecs := viper.GetInt64("jwt.id_expiration_secs")
	refExpirationSecs := viper.GetInt64("jwt.refresh_expiration_secs")

	tkConfig := &imodels.TSConfig{
		TokenRepository:       imodels.NewTokenRepo(db),
		PrivKey:               rsaPrivKey,
		PubKey:                pubKey,
		RefreshSecret:         refreshSecret,
		IDExpirationSecs:      idExpirationSecs,
		RefreshExpirationSecs: refExpirationSecs,
	}
	tokenService := imodels.NewTokenService(tkConfig)

	return &tokenService, idExpirationSecs, refExpirationSecs, nil
}

func NewTokenClient() TokenClient {
	if cfg == nil {
		cfg = services.NewConfigService(viper.ConfigFileUsed(), viper.GetString("cert.key_path"), viper.GetString("cert.cert_path"))
	}
	if fs == nil {
		fs = *services.NewFileSystemService(viper.GetString("fs.config_file_path"))
	}
	if crt == nil {
		crt = srvs.NewCertService(viper.GetString("cert.key_path"), viper.GetString("cert.cert_path"))
	}

	return &TokenClientImpl{
		cfg: cfg,
		fs:  fs,
		crt: crt,
	}
}
