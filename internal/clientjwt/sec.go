package clientjwt

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	i "github.com/faelmori/gkbxsrv/internal/models"
	"github.com/faelmori/gkbxsrv/services"
	s "github.com/faelmori/gkbxsrv/services"
	"github.com/spf13/viper"
	"log"
)

type TSConfig = i.TSConfig
type TokenService interface{ i.TokenService }
type PrivateKey = *rsa.PrivateKey
type PublicKey = *rsa.PublicKey

type TokenClient interface {
	LoadPrivateKey() (*PrivateKey, error)
	LoadPublicKey() *PublicKey
	LoadTokenCfg() (TokenService, int64, int64, error)
}
type TokenClientImpl struct {
	cfgSrv                s.ConfigService
	dbSrv                 s.DatabaseService
	fsSrv                 s.FilesystemService
	crtSrv                s.CertService
	TokenService          i.TokenService
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

func (t *TokenClientImpl) LoadPrivateKey() (*PrivateKey, error) {
	keyDataB64 := viper.GetString("jwt.private_key")
	if keyDataB64 == "" {
		return nil, fmt.Errorf("error reading private key file: %v", "jwt.private_key is empty")
	}
	keyData, err := base64.StdEncoding.DecodeString(keyDataB64)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v", err)
	}
	pwd, pwdErr := t.crtSrv.RetrievePassword()
	if pwdErr != nil {
		return nil, fmt.Errorf("error reading private key file: %v", pwdErr)
	}
	privateKey, privateKeyErr := t.crtSrv.DecryptPrivateKey(keyData, []byte(pwd))
	if privateKeyErr != nil {
		return nil, fmt.Errorf("error parsing private key: %v", privateKeyErr)
	}
	return &privateKey, nil
}
func (t *TokenClientImpl) LoadPublicKey() *PublicKey {
	if readErr := viper.ReadInConfig(); readErr != nil {
		log.Fatalf("Error reading public key file: %v", readErr)
	}
	if !t.fsSrv.ExistsConfigFile() {
		log.Fatalf("Error reading public key file: %v", "config file not found")
	}

	if !t.cfgSrv.IsConfigLoaded() {
		if loadErr := t.cfgSrv.LoadConfig(); loadErr != nil {
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

	return &publicKey
}
func (t *TokenClientImpl) LoadTokenCfg() (TokenService, int64, int64, error) {
	t.fsSrv = *s.NewFileSystemService(t.cfgSrv.GetConfigPath())
	t.crtSrv = s.NewCertService(t.fsSrv.GetDefaultKeyPath(), t.fsSrv.GetDefaultCertPath())
	db, dbErr := t.dbSrv.GetDB()
	if dbErr != nil {
		return nil, 0, 0, dbErr
	}
	tokenRepo := i.NewTokenRepo(db)
	privKey := viper.GetString("jwt.private_key")
	if privKey == "" {
		return nil, 0, 0, fmt.Errorf("error reading private key file: %v", "jwt.private_key is empty")
	}
	rsaPrivKey, rsaPrivKeyErr := t.LoadPrivateKey()
	if rsaPrivKeyErr != nil {
		return nil, 0, 0, rsaPrivKeyErr
	}
	rsaPrvK := *rsaPrivKey
	if validateErr := rsaPrvK.Validate(); validateErr != nil {
		return nil, 0, 0, fmt.Errorf("error reading private key file: %v", validateErr)
	}
	pubKey := &rsa.PublicKey{
		N: rsaPrvK.N,
		E: rsaPrvK.E,
	}
	if !pubKey.Equal(rsaPrvK.Public()) {
		return nil, 0, 0, fmt.Errorf("error reading private key file: %v", "public key does not match private key")
	}

	refreshSecret := viper.GetString("jwt.refresh_secret")
	idExpirationSecs := viper.GetInt64("jwt.id_expiration_secs")
	refExpirationSecs := viper.GetInt64("jwt.refresh_expiration_secs")

	tkConfig := &TSConfig{
		TokenRepository:       tokenRepo,
		PrivKey:               rsaPrvK,
		PubKey:                pubKey,
		RefreshSecret:         refreshSecret,
		IDExpirationSecs:      idExpirationSecs,
		RefreshExpirationSecs: refExpirationSecs,
	}
	tokenService := i.NewTokenService(tkConfig)

	return tokenService, idExpirationSecs, refExpirationSecs, nil
}

func NewTokenClient(cfg s.ConfigService, fs s.FilesystemService, crt s.CertService, db s.DatabaseService) TokenClient {
	if cfg == nil {
		cfg = services.NewConfigService(viper.ConfigFileUsed(), viper.GetString("cert.key_path"), viper.GetString("cert.cert_path"))
	}
	if fs == nil {
		fs = *services.NewFileSystemService(viper.GetString("fs.config_file_path"))
	}
	if crt == nil {
		crt = s.NewCertService(viper.GetString("cert.key_path"), viper.GetString("cert.cert_path"))
	}
	if db == nil {
		db = s.NewDatabaseService(viper.GetString("db.connection_string"))
	}
	tokenClient := &TokenClientImpl{
		cfgSrv: cfg,
		fsSrv:  fs,
		crtSrv: crt,
		dbSrv:  db,
	}

	return tokenClient
}
