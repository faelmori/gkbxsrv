package services

import (
	dbAbs "github.com/faelmori/gkbxsrv/internal/services"
)

type DatabaseService = dbAbs.IDatabaseService

func NewDatabaseService(configFile string) DatabaseService {
	return dbAbs.NewDatabaseService(configFile)
}
