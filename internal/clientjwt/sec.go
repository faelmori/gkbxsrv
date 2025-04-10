package clientjwt

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	kbxApi "github.com/faelmori/kbxutils/utils/interfaces"
	"github.com/golang-jwt/jwt/v5"

	m "github.com/faelmori/gkbxsrv/internal/models"
	a "github.com/faelmori/kbxutils/factory"
	v "github.com/spf13/viper"

	"log"
)

type TSConfig = m.TSConfig
type TokenService interface{ m.TokenService }
type PrivateKey = *rsa.PrivateKey
type PublicKey = *rsa.PublicKey

type TokenClient interface {
	LoadPrivateKey() (*PrivateKey, error)
	LoadPublicKey() *PublicKey
	LoadTokenCfg() (TokenService, int64, int64, error)
}
type TokenClientImpl struct {
	cfgSrv                kbxApi.ConfigService
	dbSrv                 kbxApi.IDatabaseService
	fsSrv                 kbxApi.FileSystemService
	crtSrv                kbxApi.ICertService
	TokenService          m.TokenService
	IDExpirationSecs      int64
	RefreshExpirationSecs int64
}

func (t *TokenClientImpl) LoadPrivateKey() (*PrivateKey, error) {
	keyDataB64 := v.GetString("jwt.private_key")
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
	if readErr := v.ReadInConfig(); readErr != nil {
		log.Fatalf("Error reading public key file: %v", readErr)
	}
	if !t.fsSrv.ExistsConfigFile() {
		log.Fatalf("Error reading public key file: %v", "config file not found")
	}

	if !t.cfgSrv.IsConfigLoaded() {
		//if loadErr := t.cfgSrv.LoadConfig(); loadErr != nil {
		//	log.Fatalf("Error loading config file: %v", loadErr)
		//}
	}

	keyDataB64 := v.GetString("jwt.public_key")
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
	t.fsSrv = *a.NewFilesystemService(t.cfgSrv.GetConfigPath())
	t.crtSrv = a.NewCertService(t.fsSrv.GetDefaultKeyPath(), t.fsSrv.GetDefaultCertPath())
	db, dbErr := t.dbSrv.GetDB()
	if dbErr != nil {
		return nil, 0, 0, dbErr
	}
	tokenRepo := m.NewTokenRepo(db)
	privKey := v.GetString("jwt.private_key")
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

	refreshSecret := v.GetString("jwt.refresh_secret")
	idExpirationSecs := v.GetInt64("jwt.id_expiration_secs")
	refExpirationSecs := v.GetInt64("jwt.refresh_expiration_secs")

	tkConfig := &TSConfig{
		TokenRepository:       tokenRepo,
		PrivKey:               rsaPrvK,
		PubKey:                pubKey,
		RefreshSecret:         refreshSecret,
		IDExpirationSecs:      idExpirationSecs,
		RefreshExpirationSecs: refExpirationSecs,
	}
	tokenService := m.NewTokenService(tkConfig)

	return tokenService, idExpirationSecs, refExpirationSecs, nil
}

func NewTokenClient(cfg kbxApi.IConfigService, fs kbxApi.FileSystemService, crt kbxApi.ICertService, db kbxApi.IDatabaseService) TokenClient {
	if cfg == nil {
		cfg = a.NewConfigService(v.ConfigFileUsed(), v.GetString("cert.key_path"), v.GetString("cert.cert_path"))
	}
	if fs == nil {
		fs = *a.NewFilesystemService(v.GetString("fs.config_file_path"))
	}
	if crt == nil {
		crt = a.NewCertService(v.GetString("cert.key_path"), v.GetString("cert.cert_path"))
	}
	if db == nil {
		db = a.NewDatabaseService(v.GetString("db.connection_string"))
	}
	tokenClient := &TokenClientImpl{
		cfgSrv: nil,
		fsSrv:  fs,
		crtSrv: crt,
		dbSrv:  db,
	}

	return tokenClient
}
