package services

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"github.com/faelmori/gokubexfs/internal/globals"
	"github.com/faelmori/gokubexfs/internal/services/filesystem"
	"github.com/faelmori/kbx/mods/logz"
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
	"os/exec"
	"strings"
	"sync"
	"time"
)

var databaseService *DatabaseServiceImpl

type DatabaseService interface {
	LoadViperConfig()
	GetDBConfig(name string) (Database, error)
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
	fs        filesystem.FileSystemService
	db        *gorm.DB
	mtx       *sync.Mutex
	wg        *sync.WaitGroup
	dbCfg     Database
	dbChanCtl chan string
	dbChanErr chan error
	dbChanSig chan os.Signal
	mdRepo    *globals.GenericRepo
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
		_ = logz.ErrorLog(fmt.Sprintf("❌ Banco de dados não conectado. Tentando reconectar..."), "GoKubexFS")
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
	_ = logz.InfoLog("Reconnecting to database...", "GDBase", logz.QUIET)
	d.dbChanCtl <- "reconnect"
	_ = logz.InfoLog("Database reconnection successful!", "GDBase", logz.QUIET)
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
		cfg := NewConfigService(d.fs.GetConfigFilePath(), d.fs.GetDefaultKeyPath(), d.fs.GetDefaultCertPath())
		dbCfg := cfg.GetDatabaseConfig()
		d = &DatabaseServiceImpl{
			fs:        filesystem.NewFileSystemSrv(""),
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
func (d *DatabaseServiceImpl) GetDBConfig(name string) (Database, error) {
	d.LoadViperConfig()
	return d.dbCfg, nil
}

func (d *DatabaseServiceImpl) GetHost() (string, error) {
	_ = logz.InfoLog("Getting database host...", "GDBase", logz.QUIET)
	d.dbChanCtl <- "getHost"
	err := <-d.dbChanErr
	_ = logz.InfoLog(fmt.Sprintf("Database host: %s", d.dbCfg.Host), "GDBase", logz.QUIET)
	if err != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Failed to get database host: %v", err), "GDBase", logz.QUIET)
		return "", err
	}
	_ = logz.InfoLog(fmt.Sprintf("Successfully got database host: %s", d.dbCfg.Host), "GDBase", logz.QUIET)
	return d.dbCfg.Host, nil
}
func (d *DatabaseServiceImpl) GetConnection(client string) (*gorm.DB, error) {
	d.LoadViperConfig()
	return d.db, nil
}
func (d *DatabaseServiceImpl) OpenDB() (*gorm.DB, error) {
	d.LoadViperConfig()
	if d.db != nil {
		_ = logz.InfoLog("Database connection already open", "GDBase", logz.QUIET)
		return d.db, nil
	}
	var db *gorm.DB
	var dbErr error
	_ = logz.InfoLog("Opening database connection...", "GDBase", logz.QUIET)
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
		_ = logz.WarnLog(fmt.Sprintf("Database type not specified or invalid (%s), falling back to SQLite", d.dbCfg.Type), "GDBase", logz.QUIET)
		db, dbErr = d.ConnectSQLite()
	}
	if dbErr != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("Failed to connect to database: %v", dbErr), "GDBase", logz.QUIET)
	}

	// The idea is to migrate the models to the database, if we get the models from the models files in the cache (generated by the gdbase models -a "hash of the models" command). This is a work in progress.
	hashFromConfigFile := md5.New().Sum([]byte(fmt.Sprintf("%s:%s:%s:%s:%s", d.dbCfg.Host, d.dbCfg.Port, d.dbCfg.Username, d.dbCfg.Password, d.dbCfg.Name)))
	cmdGenModels := exec.Command("gdbase", "models", "-a", string(hashFromConfigFile))
	cmdGenModelsErr := cmdGenModels.Run()
	if cmdGenModelsErr != nil {
		_ = logz.WarnLog(fmt.Sprintf("Failed to generate models from cache: %v", cmdGenModelsErr), "GDBase", logz.QUIET)
	}
	output, outputErr := cmdGenModels.Output()
	if outputErr != nil {
		_ = logz.WarnLog(fmt.Sprintf("Failed to get output from command: %v", outputErr), "GDBase", logz.QUIET)
	}
	outputModelsPathList := strings.Split(string(output), "\n")
	if len(outputModelsPathList) > 0 {
		_ = logz.InfoLog(fmt.Sprintf("Trying to migrate models to %s database...", d.dbCfg.Type), "GDBase", logz.QUIET)
		dinLoadedModels, dinLoadedModelsErr := d.loadModelsFromCache(outputModelsPathList)
		if dinLoadedModelsErr != nil {
			_ = logz.ErrorLog(fmt.Sprintf("Failed to load models from cache: %v", dinLoadedModelsErr), "GDBase", logz.QUIET)
			return nil, dinLoadedModelsErr
		}
		if migrateErr := d.db.AutoMigrate(dinLoadedModels...); migrateErr != nil {
			_ = logz.ErrorLog(fmt.Sprintf("Failed to migrate models to %s database: %v", d.dbCfg.Type, migrateErr), "GDBase", logz.QUIET)
			return nil, migrateErr
		}
		_ = logz.InfoLog(fmt.Sprintf("Successfully migrated models to %s database", d.dbCfg.Type), "GDBase", logz.QUIET)
	} else {
		_ = logz.WarnLog(fmt.Sprintf("Models not found in cache, skipping migration"), "GDBase", logz.QUIET)
	}

	_ = logz.InfoLog(fmt.Sprintf("Successfully migrated models to %s database", d.dbCfg.Type), "GDBase", logz.QUIET)
	return db, nil
}
func (d *DatabaseServiceImpl) loadModelsFromCache(generatedFilesPath []string) ([]interface{}, error) {
	var modelsList []interface{}
	for _, modelPath := range generatedFilesPath {
		_ = logz.InfoLog(fmt.Sprintf("Loading model from JSON: %s", modelPath), "GDBase", logz.QUIET)
		readFile, readFileErr := os.ReadFile(modelPath)
		if readFileErr != nil {
			_ = logz.ErrorLog(fmt.Sprintf("Failed to read model file: %v", readFileErr), "GDBase", logz.QUIET)
			return nil, readFileErr
		}
		jsonContent, jsonContentErr := json.Marshal(readFile)
		if jsonContentErr != nil {
			_ = logz.ErrorLog(fmt.Sprintf("Failed to marshal JSON content: %v", jsonContentErr), "GDBase", logz.QUIET)
			return nil, jsonContentErr
		}
		jsonModelStructRef, jsonModelStructRefErr := d.loadModelFromJSONContent(jsonContent)
		if jsonModelStructRefErr != nil {
			_ = logz.ErrorLog(fmt.Sprintf("Failed to load model from JSON: %v", jsonModelStructRefErr), "GDBase", logz.QUIET)
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
			_ = logz.ErrorLog(fmt.Sprintf("Failed to unmarshal JSON content: %v", jsonUnmarshalErr), "GDBase", logz.QUIET)
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
		_ = logz.WarnLog(fmt.Sprintf("Initial error connecting to database: %v", dbErr), "GDBase", logz.QUIET)
		return d.InitialHealthCheck()
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) ConnectPostgres() (*gorm.DB, error) {
	d.LoadViperConfig()
	_ = logz.InfoLog("Connecting to Postgres database...", "GDBase", logz.QUIET)
	if d.db == nil {
		_ = logz.InfoLog("Database connection is nil, creating new connection...", "GDBase", logz.QUIET)
		var dsn string
		if d.dbCfg.Dsn != "" {
			_ = logz.InfoLog("Using DSN from config file...", "GDBase", logz.QUIET)
			dsn = d.dbCfg.Dsn
		} else {
			_ = logz.InfoLog("DSN not found in config file, creating new DSN...", "GDBase", logz.QUIET)
			dsn = fmt.Sprintf(
				"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo TLS=false",
				d.dbCfg.Host, d.dbCfg.Port, d.dbCfg.Username, d.dbCfg.Password, d.dbCfg.Name,
			)
		}
		d.dbCfg.Dsn = dsn
		var dbErr error
		_ = logz.InfoLog("Opening new connection to Postgres database...", "GDBase", logz.QUIET)
		d.db, dbErr = gorm.Open(postgres.Open(d.dbCfg.Dsn), &gorm.Config{})
		if dbErr != nil {
			_ = logz.WarnLog(fmt.Sprintf("Initial error connecting to database: %v", dbErr), "GDBase", logz.QUIET)
			return d.InitialHealthCheck()
		}
	}
	_ = logz.InfoLog("Checking database health...", "GDBase", logz.QUIET)
	return d.db, nil
}
func (d *DatabaseServiceImpl) ConnectSQLite() (*gorm.DB, error) {
	d.LoadViperConfig()
	var dbErr error
	d.db, dbErr = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if dbErr != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("Initial error connecting to database: %v", dbErr), "GDBase", logz.QUIET)
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
		_ = logz.WarnLog(fmt.Sprintf("Initial error connecting to database: %v", dbErr), "GDBase", logz.QUIET)
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
			return logz.ErrorLog(fmt.Sprintf("❌ Database connection is nil"), "GDBase", logz.QUIET)
		}
		return d.db.Raw("SELECT 1").Error
	}
	return logz.ErrorLog(fmt.Sprintf("❌ Database connection is nil"), "GDBase", logz.QUIET)
}
func (d *DatabaseServiceImpl) WaitForDatabase(timeout time.Duration, maxRetries int) error {
	d.LoadViperConfig()
	_ = logz.InfoLog("Checking database health...", "GDBase", logz.QUIET)
	retryInterval := timeout
	for i := 0; i < maxRetries; i++ {
		if err := d.CheckDatabaseHealth(); err == nil {
			return nil
		} else {
			_ = logz.ErrorLog(fmt.Sprintf("Erro ao verificar a conexão com o banco de dados: %v", err), "GDBase", logz.QUIET)
		}
		time.Sleep(retryInterval)
	}
	return logz.ErrorLog(fmt.Sprintf("Falha na conexão com o banco de dados após %d tentativas", maxRetries), "GDBase", logz.QUIET)
}
func (d *DatabaseServiceImpl) InitialHealthCheck() (*gorm.DB, error) {
	d.LoadViperConfig()
	_ = logz.InfoLog("Checking database health...", "GDBase", logz.QUIET)
	timeout := 2 * time.Second
	maxRetries := 3
	if waitErr := d.WaitForDatabase(timeout, maxRetries); waitErr != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("Failed to check database health: %v", waitErr), "GDBase", logz.QUIET)
	} else {
		_ = logz.InfoLog("Database health check successful!", "GDBase", logz.QUIET)
		return d.db, nil
	}
}
func (d *DatabaseServiceImpl) SetConnection(db *gorm.DB) {
	d.LoadViperConfig()
	if db != nil {
		_ = logz.InfoLog(fmt.Sprintf("Setting database connection (%s)...", db.Name()), "GDBase")
		d.db = db
		_ = logz.InfoLog("Database connection set successfully!", "GDBase")
	} else {
		_ = logz.WarnLog("Database connection is nil", "GDBase")
	}
}

func (d *DatabaseServiceImpl) ServiceHandler() chan interface{} {
	d.LoadViperConfig()
	for {
		_ = logz.InfoLog(fmt.Sprintf("🔁 Aguardando dados do canal..."), "GoKubexFS", logz.QUIET)
		select {
		case dbCtl := <-d.dbChanCtl:
			// d.mtx.Lock()
			switch dbCtl {
			case "reconnect":
				_ = logz.InfoLog(fmt.Sprintf("🔄 Reconectando ao banco de dados..."), "GoKubexFS", logz.QUIET)
				db := *d.db
				dbClientB, dbClientBErr := db.DB()
				if dbClientBErr != nil {
					_ = logz.ErrorLog(fmt.Sprintf("❌ Erro ao obter a conexão SQL: %v", dbClientBErr), "GoKubexFS")
					d.dbChanErr <- &globals.ValidationError{Message: dbClientBErr.Error(), Field: "dbChanCtl_reconnect"}
				} else {
					if dbClientBErrB := dbClientB.Close(); dbClientBErrB != nil {
						_ = logz.ErrorLog(fmt.Sprintf("❌ Erro ao fechar a conexão SQL: %v", dbClientBErrB), "GoKubexFS")
						d.dbChanErr <- &globals.ValidationError{Message: dbClientBErrB.Error(), Field: "dbChanCtl_reconnect"}
					} else {
						_ = logz.InfoLog(fmt.Sprintf("✅ Reconexão ao banco de dados realizada com sucesso!"), "GoKubexFS", logz.QUIET)
						d.dbChanErr <- nil
					}
				}
				_ = logz.InfoLog(fmt.Sprintf("✅ Reconexão ao banco de dados realizada com sucesso!"), "GoKubexFS", logz.QUIET)
			case "isConnected":
				_ = logz.InfoLog(fmt.Sprintf("🔗 Verificando conexão ao banco de dados..."), "GoKubexFS", logz.QUIET)
				if _, dbErr := d.db.DB(); dbErr != nil {
					_ = logz.ErrorLog(fmt.Sprintf("❌ Erro ao obter a conexão SQL: %v", dbErr), "GoKubexFS")
					d.dbChanErr <- &globals.ValidationError{Message: dbErr.Error(), Field: "dbChanCtl_isConnected"}
				} else {
					_ = logz.InfoLog(fmt.Sprintf("✅ Conexão ao banco de dados realizada com sucesso!"), "GoKubexFS", logz.QUIET)
					d.dbChanErr <- nil
				}
				_ = logz.InfoLog(fmt.Sprintf("✅ Conexão ao banco de dados verificada com sucesso!"), "GoKubexFS", logz.QUIET)
			case "getHost":
				_ = logz.InfoLog(fmt.Sprintf("🔗 Obtendo host do banco de dados..."), "GoKubexFS", logz.QUIET)
				dbCfgB := &d.dbCfg
				_ = logz.InfoLog(fmt.Sprintf("✅ Host do banco de dados: %s", dbCfgB.Host), "GoKubexFS", logz.QUIET)
				d.dbChanErr <- &globals.ValidationError{Message: dbCfgB.Host, Field: "dbChanCtl_getHost"}
				_ = logz.InfoLog(fmt.Sprintf("✅ Host do banco de dados obtido com sucesso!"), "GoKubexFS", logz.QUIET)
			case "close":
				_ = logz.InfoLog(fmt.Sprintf("🔗 Fechando conexão ao banco de dados..."), "GoKubexFS", logz.QUIET)
				dbB, dbBErr := d.db.DB()
				if dbBErr != nil {
					_ = logz.ErrorLog(fmt.Sprintf("❌ Erro ao obter a conexão SQL: %v", dbBErr), "GoKubexFS")
					d.dbChanErr <- &globals.ValidationError{Message: dbBErr.Error(), Field: "dbChanCtl_close"}
				} else {
					_ = logz.InfoLog(fmt.Sprintf("🔗 Fechando conexão ao banco de dados..."), "GoKubexFS", logz.QUIET)
					_ = dbB.Close()
					d.dbChanErr <- nil
				}
				_ = logz.InfoLog(fmt.Sprintf("✅ Conexão ao banco de dados fechada com sucesso!"), "GoKubexFS", logz.QUIET)
			}
			// d.mtx.Unlock()
		}
	}
}

func NewDatabaseService(configFileArg string) DatabaseService {
	if databaseService != nil {
		return databaseService
	}
	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando serviço de banco de dados..."), "GoKubexFS", logz.QUIET)
	cfg := NewConfigService(configFileArg, "", "")
	cfgLoadErr := cfg.LoadConfig()
	if cfgLoadErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("❌ Erro ao carregar o arquivo de configuração: %v", cfgLoadErr), "GoKubexFS")
		os.Exit(1)
	}
	_ = logz.InfoLog(fmt.Sprintf("🔗 Configurações do banco de dados: %v", cfg.GetConfigPath()), "GoKubexFS", logz.QUIET)
	databaseService = &DatabaseServiceImpl{
		fs:        Fs,
		db:        nil,
		mtx:       &sync.Mutex{},
		wg:        &sync.WaitGroup{},
		dbCfg:     Database{},
		dbChanCtl: make(chan string, 10),
		dbChanErr: make(chan error, 10),
		dbChanSig: make(chan os.Signal, 1),
	}
	databaseService.LoadViperConfig()
	_ = logz.InfoLog(fmt.Sprintf("✅ Serviço de banco de dados iniciado com sucesso!"), "GoKubexFS", logz.QUIET)
	db, dbErr := databaseService.OpenDB()
	if dbErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("❌ Erro ao conectar ao banco de dados: %v", dbErr), "GoKubexFS")
	} else {
		_ = logz.InfoLog(fmt.Sprintf("✅ Conexão ao banco de dados realizada com sucesso!"), "GoKubexFS", logz.QUIET)
		databaseService.db = db
	}
	_ = logz.InfoLog(fmt.Sprintf("🔗 Iniciando handler de serviços..."), "GoKubexFS", logz.QUIET)
	return databaseService
}

func contains(output, name string) bool {
	for _, line := range splitLines(output) {
		if line == name {
			return true
		}
	}
	return false
}
func splitLines(output string) []string {
	return strings.Split(output, "\n")
}
