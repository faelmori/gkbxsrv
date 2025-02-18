package services

import (
	. "github.com/faelmori/gokubexfs/internal/services/filesystem"
)

type FilesystemService interface{ FileSystemService }

func NewFileSystemService(configFile string) *FilesystemService {
	srv := NewFileSystemSrv(configFile)
	srvB := srv.(FilesystemService)
	return &srvB
}
