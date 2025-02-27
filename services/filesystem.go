package services

import (
	fsys "github.com/faelmori/gkbxsrv/internal/services"
)

type FilesystemService interface{ fsys.FileSystemService }

func NewFileSystemService(configFile string) *FilesystemService {
	srv := fsys.NewFileSystemSrv(configFile)
	srvB := srv.(FilesystemService)
	return &srvB
}
