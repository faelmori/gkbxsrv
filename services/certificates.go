package services

import (
	cert "github.com/faelmori/gkbxsrv/internal/services"
)

type CertService = cert.ICertService

func NewCertService(keyPath string, certPath string) *CertService {
	srv := cert.NewCertService(keyPath, certPath)
	srvB := srv.(CertService)
	return &srvB
}
