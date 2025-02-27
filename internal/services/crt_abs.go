package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/globals"
	"github.com/faelmori/gkbxsrv/internal/utils"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/chacha20poly1305"
	"math/big"
	"os"
	"strings"
	"time"
)

type ICertService interface {
	StorePassword(password string) error
	RetrievePassword() (string, error)
	GenerateRandomKey(length int) (string, error)
	GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, error)
	GenSelfCert() ([]byte, []byte, error)
	DecryptPrivateKey(ciphertext []byte, password []byte) (*rsa.PrivateKey, error)
	VerifyCert() error
	GetCertAndKeyFromFile() ([]byte, []byte, error)
}
type CertImpl struct {
	KeyPath  string
	CertPath string
	Keyring  string
}

func (c *CertImpl) StorePassword(password string) error {
	return keyring.Set(globals.KeyringService, globals.KeyringKey, password)
}
func (c *CertImpl) RetrievePassword() (string, error) {
	return keyring.Get(globals.KeyringService, globals.KeyringKey)
}
func (c *CertImpl) GenerateRandomKey(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var password []byte
	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("erro ao gerar índice aleatório: %w", err)
		}
		password = append(password, charset[randomIndex.Int64()])
	}
	return string(password), nil
	//key := make([]byte, 16)
	//if _, err := rand.Read(key); err != nil {
	//	return "", fmt.Errorf("erro ao gerar chave: %w", err)
	//}
	//return string(key), nil
}
func (c *CertImpl) GenerateCertificate(certPath, keyPath string, password []byte) ([]byte, []byte, error) {
	priv, generateKeyErr := rsa.GenerateKey(rand.Reader, 4096)
	if generateKeyErr != nil {
		return nil, nil, fmt.Errorf("erro ao gerar chave privada: %v", generateKeyErr)
	}

	sn, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	template := x509.Certificate{
		SerialNumber: sn,
		Subject:      pkix.Name{CommonName: "Kubex Self-Signed"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         true,
	}

	certDER, certDERErr := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if certDERErr != nil {
		return nil, nil, fmt.Errorf("erro ao criar certificado: %v", certDERErr)
	}

	pkcs1PrivBytes := x509.MarshalPKCS1PrivateKey(priv)

	block, err := chacha20poly1305.New(password)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao criar cipher: %w", err)
	}

	nonce := make([]byte, block.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("erro ao gerar nonce: %w", err)
	}

	ciphertext := block.Seal(nonce, nonce, pkcs1PrivBytes, nil)

	certFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao abrir arquivo de certificado: %w", err)
	}
	defer func(certFile *os.File) {
		_ = certFile.Close()
	}(certFile)

	pemBlock := &pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	if err := pem.Encode(certFile, pemBlock); err != nil {
		return nil, nil, fmt.Errorf("erro ao codificar certificado: %w", err)
	}

	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao abrir arquivo de chave: %w", err)
	}
	defer func(keyFile *os.File) {
		_ = keyFile.Close()
	}(keyFile)

	pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: ciphertext}
	if err := pem.Encode(keyFile, pemBlock); err != nil {
		return nil, nil, fmt.Errorf("erro ao codificar chave privada: %w", err)
	}

	return ciphertext, certDER, nil
}
func (c *CertImpl) GenSelfCert() ([]byte, []byte, error) {
	password := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(password); err != nil {
		return nil, nil, fmt.Errorf("erro ao gerar senha: %w", err)
	}

	if err := c.StorePassword(string(password)); err != nil {
		return nil, nil, fmt.Errorf("erro ao armazenar senha: %w", err)
	}

	return c.GenerateCertificate(c.CertPath, c.KeyPath, password)
}
func (c *CertImpl) DecryptPrivateKey(ciphertext []byte, password []byte) (*rsa.PrivateKey, error) {
	pwd, pwdErr := c.RetrievePassword()
	if pwdErr != nil {
		return nil, fmt.Errorf("erro ao recuperar senha: %w", pwdErr)
	}

	block, err := chacha20poly1305.New([]byte(pwd))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher: %w", err)
	}

	nonceSize := block.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext muito curto")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	if len(nonce) != nonceSize {
		return nil, fmt.Errorf("nonce incorreto")
	}

	pkcs1PrivBytes, err := block.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar chave privada: %w", err)
	}

	return x509.ParsePKCS1PrivateKey(pkcs1PrivBytes)
}
func (c *CertImpl) VerifyCert() error {
	certFile, err := os.Open(c.CertPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo de certificado: %w", err)
	}
	defer func(certFile *os.File) {
		_ = certFile.Close()
	}(certFile)

	certBytes, err := os.ReadFile(c.CertPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de certificado: %w", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return fmt.Errorf("erro ao decodificar certificado")
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("erro ao analisar certificado: %w", err)
	}

	return nil
}
func (c *CertImpl) GetCertAndKeyFromFile() ([]byte, []byte, error) {
	certBytes, err := os.ReadFile(c.CertPath)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao ler arquivo de certificado: %w", err)
	}

	keyBytes, err := os.ReadFile(c.KeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao ler arquivo de chave: %w", err)
	}

	return certBytes, keyBytes, nil
}

func NewCertService(keyPath, certPath string) ICertService {
	home, homeErr := utils.GetWorkDir()
	if homeErr != nil {
		//_ = logz.ErrorLog(homeErr.Error(), "CertService")
		fmt.Println(homeErr.Error())
		os.Exit(1)
	}
	if keyPath == "" {
		keyPath = strings.ReplaceAll(globals.DefaultKeyPath, "$HOME", home)
	}
	if certPath == "" {
		certPath = strings.ReplaceAll(globals.DefaultCertPath, "$HOME", home)
	}
	return &CertImpl{
		KeyPath:  keyPath,
		CertPath: certPath,
	}
}
