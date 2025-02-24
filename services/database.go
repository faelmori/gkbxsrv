package services

import (
	dbAbs "github.com/faelmori/gokubexfs/internal/services"
)

type DatabaseService = dbAbs.IDatabaseService

func NewDatabaseService(configFile string) DatabaseService {
	return dbAbs.NewDatabaseService(configFile)
}
