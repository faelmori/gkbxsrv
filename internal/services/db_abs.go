package services

import (
	"database/sql"
	"fmt"
	glb "github.com/faelmori/gkbxsrv/internal/globals"
	"github.com/faelmori/gkbxsrv/internal/models"
	"github.com/goccy/go-json"
	"github.com/godror/godror"
	dsn2 "github.com/godror/godror/dsn"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"os"
	"sync"
	"time"
)

var databaseService *DatabaseServiceImpl

type IDatabaseService interface {
	LoadViperConfig()
	GetDBConfig(name string) (glb.Database, error)
	ConnectDB() error
	GetDB() (*gorm.DB, error)
	CloseDBConnection() error
	ServiceHandler() chan interface{}
	Reconnect() error
	IsConnected() error
	GetHost() (string, error)
	GetConnection(client string) (*gorm.DB, error)
	OpenDB() (*gorm.DB, error)
	ConnectMySQL() (*gorm.DB, error)
	ConnectPostgres() (*gorm.DB, error)
	ConnectSQLite() (*gorm.DB, error)
	ConnectMSSQL() (*gorm.DB, error)
	ConnectOracle() (*gorm.DB, error)
	CheckDatabaseHealth() error
	WaitForDatabase(timeout time.Duration, maxRetries int) error
	InitialHealthCheck() (*gorm.DB, error)
	SetConnection(db *gorm.DB)
}

type DatabaseServiceImpl struct {
	fs        FileSystemService
	db        *gorm.DB
	mtx       *sync.Mutex
	wg        *sync.WaitGroup
	dbCfg     glb.Database
	dbChanCtl chan string
	dbChanErr chan error
	dbChanSig chan os.Signal
	mdRepo    *glb.GenericRepo
	dbStats   sql.DBStats
	lastStats sql.DBStats
}

func (d *DatabaseServiceImpl) LoadViperConfig() {
	vpDBCfg := viper.GetStringMapString("database")
	d.dbCfg.Host = vpDBCfg["host"]
	d.dbCfg.Port = vpDBCfg["port"]
	d.dbCfg.Username = vpDBCfg["username"]
	d.dbCfg.Password = vpDBCfg["password"]
	d.dbCfg.Name = vpDBCfg["name"]
	d.dbCfg.Type = vpDBCfg["type"]
	d.dbCfg.Dsn = vpDBCfg["dsn"]
	d.dbCfg.ConnectionString = vpDBCfg["connection_string"]
}

func (d *DatabaseServiceImpl) ConnectDB() error {
	d.LoadViperConfig()
	if d.dbCfg.Dsn == "" {
		if d.dbCfg.ConnectionString != "" {
			d.dbCfg.Dsn = d.dbCfg.ConnectionString
		} else {
			d.dbCfg.Dsn = fmt.Sprintf(
				"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo",
				d.dbCfg.Host, d.dbCfg.Port, d.dbCfg.Username, d.dbCfg.Password, d.dbCfg.Name,
			)
		}
	}
	db, err := gorm.Open(postgres.Open(d.dbCfg.Dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", err)
	}
	d.db = db
	return nil
}
func (d *DatabaseServiceImpl) GetDB() (*gorm.DB, error) {
	if d.db == nil {
		d.LoadViperConfig()
		if err := d.ConnectDB(); err != nil {
			return nil, fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", err)
		} else {
			return d.db, nil
		}
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) CloseDBConnection() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("❌ Erro ao obter a conexão SQL: %v", err)
	}
	return sqlDB.Close()
}
func (d *DatabaseServiceImpl) Reconnect() error {
	d.LoadViperConfig()
	d.dbChanCtl <- "reconnect"
	return <-d.dbChanErr
}
func (d *DatabaseServiceImpl) IsConnected() error {
	if d != nil {
		if err := d.CheckDatabaseHealth(); err != nil {
			return fmt.Errorf("❌ Erro ao verificar a conexão com o banco de dados: %v", err)
		}
		return nil
	} else {
		mtx := &sync.Mutex{}
		mtx.Lock()
		cfg := NewConfigSrv(d.fs.GetConfigFilePath(), d.fs.GetDefaultKeyPath(), d.fs.GetDefaultCertPath())
		dbCfg := cfg.GetDatabaseConfig()
		d = &DatabaseServiceImpl{
			fs:        NewFileSystemSrv(""),
			db:        nil,
			mtx:       &sync.Mutex{},
			wg:        &sync.WaitGroup{},
			dbCfg:     dbCfg,
			dbChanCtl: make(chan string),
			dbChanErr: make(chan error),
			dbChanSig: make(chan os.Signal),
			mdRepo:    nil,
			dbStats:   sql.DBStats{},
			lastStats: sql.DBStats{},
		}
		db, dbErr := d.OpenDB()
		if dbErr != nil {
			return fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", dbErr)
		}
		d.db = db
		mtx.Unlock()
		return nil
	}
}
func (d *DatabaseServiceImpl) GetDBConfig(name string) (glb.Database, error) {
	d.LoadViperConfig()
	return d.dbCfg, nil
}

func (d *DatabaseServiceImpl) GetHost() (string, error) {
	d.dbChanCtl <- "getHost"
	err := <-d.dbChanErr
	if err != nil {
		return "", err
	}
	return d.dbCfg.Host, nil
}
func (d *DatabaseServiceImpl) GetConnection(client string) (*gorm.DB, error) {
	d.LoadViperConfig()
	return d.db, nil
}
func (d *DatabaseServiceImpl) OpenDB() (*gorm.DB, error) {
	d.LoadViperConfig()
	if d.db != nil {
		return d.db, nil
	}
	var db *gorm.DB
	var dbErr error
	switch d.dbCfg.Type {
	case "oracle", "oci8", "goracle", "godror":
		db, dbErr = d.ConnectOracle()
	case "mysql", "mariadb":
		db, dbErr = d.ConnectMySQL()
	case "postgres", "postgresql":
		db, dbErr = d.ConnectPostgres()
	case "mssql", "sqlserver":
		db, dbErr = d.ConnectMSSQL()
	case "sqlite", "sqlite3":
		db, dbErr = d.ConnectSQLite()
	default:
		db, dbErr = d.ConnectSQLite()
	}
	if dbErr != nil {
		return nil, fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", dbErr)
	}

	if migrateErr := d.db.AutoMigrate(models.ModelList...); migrateErr != nil {
		os.Exit(1)
		return nil, migrateErr
	}
	return db, nil
}
func (d *DatabaseServiceImpl) loadModelsFromCache(generatedFilesPath []string) ([]interface{}, error) {
	var modelsList []interface{}
	for _, modelPath := range generatedFilesPath {
		readFile, readFileErr := os.ReadFile(modelPath)
		if readFileErr != nil {
			return nil, readFileErr
		}
		jsonContent, jsonContentErr := json.Marshal(readFile)
		if jsonContentErr != nil {
			return nil, jsonContentErr
		}
		jsonModelStructRef, jsonModelStructRefErr := d.loadModelFromJSONContent(jsonContent)
		if jsonModelStructRefErr != nil {
			return nil, jsonModelStructRefErr
		}
		modelsList = append(modelsList, jsonModelStructRef)
	}
	return modelsList, nil
}

func (d *DatabaseServiceImpl) loadModelFromJSONContent(jsonContent []byte) (interface{}, error) {
	var modelStructRef interface{}
	jsonContentStr := string(jsonContent)
	if jsonContentStr != "" {
		if jsonUnmarshalErr := json.Unmarshal([]byte(jsonContentStr), &modelStructRef); jsonUnmarshalErr != nil {
			return nil, jsonUnmarshalErr
		}
	}
	return modelStructRef, nil
}

func (d *DatabaseServiceImpl) ConnectMySQL() (*gorm.DB, error) {
	d.LoadViperConfig()
	var dsn string
	if d.dbCfg.Dsn != "" {
		dsn = d.dbCfg.Dsn
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			d.dbCfg.Username, d.dbCfg.Password, d.dbCfg.Host, d.dbCfg.Port, d.dbCfg.Name,
		)
	}
	d.dbCfg.Dsn = dsn
	var dbErr error
	d.db, dbErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if dbErr != nil {
		return d.InitialHealthCheck()
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) ConnectPostgres() (*gorm.DB, error) {
	d.LoadViperConfig()
	if d.db == nil {
		var dsn string
		if d.dbCfg.Dsn != "" {
			dsn = d.dbCfg.Dsn
		} else {
			dsn = fmt.Sprintf(
				"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo TLS=false",
				d.dbCfg.Host, d.dbCfg.Port, d.dbCfg.Username, d.dbCfg.Password, d.dbCfg.Name,
			)
		}
		d.dbCfg.Dsn = dsn
		var dbErr error
		d.db, dbErr = gorm.Open(postgres.Open(d.dbCfg.Dsn), &gorm.Config{})
		if dbErr != nil {
			return d.InitialHealthCheck()
		}
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) ConnectSQLite() (*gorm.DB, error) {
	d.LoadViperConfig()
	var dbErr error
	d.db, dbErr = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if dbErr != nil {
		return nil, fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", dbErr)
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) ConnectMSSQL() (*gorm.DB, error) {
	d.LoadViperConfig()
	var dsn string
	if d.dbCfg.Dsn != "" {
		dsn = d.dbCfg.Dsn
	} else {
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			d.dbCfg.Username, d.dbCfg.Password, d.dbCfg.Host, d.dbCfg.Port, d.dbCfg.Name,
		)
	}
	d.dbCfg.Dsn = dsn
	var dbErr error
	d.db, dbErr = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if dbErr != nil {
		return d.InitialHealthCheck()
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) ConnectOracle() (*gorm.DB, error) {
	d.LoadViperConfig()
	dsn := viper.GetString("database.connection_string")
	if dsn == "" {
		dsn = fmt.Sprintf("%s/%s@%s:%s/%s", d.dbCfg.Username, d.dbCfg.Password, d.dbCfg.Host, d.dbCfg.Port, d.dbCfg.Name)
	}
	var blSA sql.NullBool
	_ = blSA.Scan(true)
	dsnCP := dsn2.ConnectionParams{}
	cmParams := dsn2.CommonParams{}
	dsnCP.Username = d.dbCfg.Username
	dsnCP.Password = dsn2.NewPassword(d.dbCfg.Password)
	dsnCP.Timezone = time.Local
	dsnCP.StandaloneConnection = blSA
	cmParams.ConnectString = fmt.Sprintf("%s:%s/%s", d.dbCfg.Host, d.dbCfg.Port, d.dbCfg.Name)
	dsnCP.CommonParams = cmParams
	oracle := godror.NewConnector(dsnCP)
	oraConn, oraConnErr := oracle.Connect(context.Background())
	if oraConnErr != nil {
		return nil, oraConnErr
	}
	var dbErr error
	d.db, dbErr = gorm.Open(oraConn.(gorm.Dialector), &gorm.Config{})
	if dbErr != nil {
		return d.InitialHealthCheck()
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) CheckDatabaseHealth() error {
	d.LoadViperConfig()
	if d != nil {
		d.LoadViperConfig()
		if d.db == nil {
			return fmt.Errorf("❌ Database connection is nil")
		}
		return d.db.Raw("SELECT 1").Error
	}
	return fmt.Errorf("❌ Database connection is nil")
}
func (d *DatabaseServiceImpl) WaitForDatabase(timeout time.Duration, maxRetries int) error {
	d.LoadViperConfig()
	retryInterval := timeout
	for i := 0; i < maxRetries; i++ {
		if err := d.CheckDatabaseHealth(); err == nil {
			return nil
		} else {
			fmt.Printf("Erro ao verificar a conexão com o banco de dados: %v", err)
		}
		time.Sleep(retryInterval)
	}
	return fmt.Errorf("Falha na conexão com o banco de dados após %d tentativas", maxRetries)
}
func (d *DatabaseServiceImpl) InitialHealthCheck() (*gorm.DB, error) {
	d.LoadViperConfig()
	timeout := 2 * time.Second
	maxRetries := 3
	if waitErr := d.WaitForDatabase(timeout, maxRetries); waitErr != nil {
		return nil, fmt.Errorf("❌ Erro ao verificar a saúde do banco de dados: %v", waitErr)
	} else {
		return d.db, nil
	}
}
func (d *DatabaseServiceImpl) SetConnection(db *gorm.DB) {
	d.LoadViperConfig()
	if db != nil {
		d.db = db
	} else {
		fmt.Println("Database connection is nil")
	}
}

func (d *DatabaseServiceImpl) ServiceHandler() chan interface{} {
	d.LoadViperConfig()
	for {
		select {
		case dbCtl := <-d.dbChanCtl:
			switch dbCtl {
			case "reconnect":
				db := *d.db
				dbClientB, dbClientBErr := db.DB()
				if dbClientBErr != nil {
					d.dbChanErr <- &glb.ValidationError{Message: dbClientBErr.Error(), Field: "dbChanCtl_reconnect"}
				} else {
					if dbClientBErrB := dbClientB.Close(); dbClientBErrB != nil {
						d.dbChanErr <- &glb.ValidationError{Message: dbClientBErrB.Error(), Field: "dbChanCtl_reconnect"}
					} else {
						d.dbChanErr <- nil
					}
				}
			case "isConnected":
				if _, dbErr := d.db.DB(); dbErr != nil {
					d.dbChanErr <- &glb.ValidationError{Message: dbErr.Error(), Field: "dbChanCtl_isConnected"}
				} else {
					d.dbChanErr <- nil
				}
			case "getHost":
				dbCfgB := &d.dbCfg
				d.dbChanErr <- &glb.ValidationError{Message: dbCfgB.Host, Field: "dbChanCtl_getHost"}
			case "close":
				dbB, dbBErr := d.db.DB()
				if dbBErr != nil {
					d.dbChanErr <- &glb.ValidationError{Message: dbBErr.Error(), Field: "dbChanCtl_close"}
				} else {
					_ = dbB.Close()
					d.dbChanErr <- nil
				}
			}
		}
	}
}

func NewDatabaseService(configFileArg string) IDatabaseService {
	if databaseService != nil {
		return databaseService
	}
	cfg := NewConfigSrv(configFileArg, "", "")
	cfgLoadErr := cfg.LoadConfig()
	if cfgLoadErr != nil {
		os.Exit(1)
	}
	databaseService = &DatabaseServiceImpl{
		fs:        NewFileSystemSrv(""),
		db:        nil,
		mtx:       &sync.Mutex{},
		wg:        &sync.WaitGroup{},
		dbCfg:     glb.Database{},
		dbChanCtl: make(chan string, 10),
		dbChanErr: make(chan error, 10),
		dbChanSig: make(chan os.Signal, 1),
	}
	databaseService.LoadViperConfig()
	db, dbErr := databaseService.OpenDB()
	if dbErr != nil {
	} else {
		databaseService.db = db
	}
	return databaseService
}
