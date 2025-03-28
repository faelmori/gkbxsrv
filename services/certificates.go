package services

import (
	cert "github.com/faelmori/gkbxsrv/internal/services"
)

type CertService interface {
	cert.ICertService
}

func NewCertService(keyPath string, certPath string) CertService {
	return cert.NewCertService(keyPath, certPath)
}
