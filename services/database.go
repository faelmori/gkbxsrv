package services

import (
	"database/sql"
	"fmt"
	dbModels "github.com/faelmori/gdbase/models"
	"github.com/faelmori/gokubexfs/internal/globals"
	"github.com/faelmori/kbx/mods/logz"
	"github.com/godror/godror"
	dsn2 "github.com/godror/godror/dsn"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var databaseService *DatabaseServiceImpl

type DatabaseService interface {
	IsDockerRunning() bool
	IsServiceRunning(serviceName string) bool
	ExistsContainer(containerName string) bool
	StartService(serviceName, image string, ports []string, envVars []string)
	SetupDatabaseServices() error
	StartMongoDB()
	StartPostgres()
	StartRabbitMQ()
	StartRedis()
	GetDBConfig(name string) (Database, error)
	ConnectDB() error
	GetDB() (*gorm.DB, error)
	CloseDBConnection() error
	ServiceHandler(dbChanData <-chan interface{})
	Reconnect() error
	IsConnected() error
	GetHost() (string, error)
	GetConnection(client string) (*gorm.DB, error)
	OpenDB() (*gorm.DB, error)
	connectMySQL() (*gorm.DB, error)
	connectPostgres() (*gorm.DB, error)
	connectSQLite() (*gorm.DB, error)
	connectMSSQL() (*gorm.DB, error)
	connectOracle() (*gorm.DB, error)
	checkDatabaseHealth() error
	waitForDatabase(timeout time.Duration, maxRetries int) error
	initialHealthCheck() (*gorm.DB, error)
	SetConnection(db *gorm.DB)
}

type DatabaseServiceImpl struct {
	fs        FileSystemService
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

func (d *DatabaseServiceImpl) loadViperConfig() {
	//mtx := &sync.Mutex{}
	//mtx.Lock()
	//cfg := NewConfigService(d.fs.GetConfigFilePath(), d.fs.GetDefaultKeyPath(), d.fs.GetDefaultCertPath())
	//dbCfg := cfg.GetDatabaseConfig()
	//d = &DatabaseServiceImpl{
	//	fs:        NewFileSystemService(""),
	//	db:        nil,
	//	mtx:       &sync.Mutex{},
	//	wg:        &sync.WaitGroup{},
	//	dbCfg:     dbCfg,
	//	dbChanCtl: make(chan string),
	//	dbChanErr: make(chan error),
	//	dbChanSig: make(chan os.Signal),
	//	mdRepo:    nil,
	//	dbStats:   sql.DBStats{},
	//	lastStats: sql.DBStats{},
	//}
	//db, dbErr := d.OpenDB()
	//if dbErr != nil {
	//	return fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", dbErr)
	//}
	//d.db = db
	//mtx.Unlock()

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
func (d *DatabaseServiceImpl) IsDockerRunning() bool {
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Docker não está rodando: %v", err)
	}
	return true
}
func (d *DatabaseServiceImpl) IsServiceRunning(serviceName string) bool {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", serviceName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		_ = logz.InfoLog(fmt.Sprintf("❌ Erro ao verificar serviço Docker: %v", err), "GoKubexFS", logz.QUIET)
	}
	return string(output) != ""
}
func (d *DatabaseServiceImpl) ExistsContainer(containerName string) bool {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		_ = logz.InfoLog(fmt.Sprintf("❌ Erro ao verificar containers: %v", err), "GoKubexFS", logz.QUIET)
	}
	return contains(string(output), containerName)
}
func (d *DatabaseServiceImpl) StartMongoDB() {
	mongoPort := viper.GetString("mongodb.port")

	if d.IsServiceRunning("kubex-mongodb") {
		//
		_ = logz.InfoLog(fmt.Sprintf("✅ MongoDB já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando MongoDB..."), "GoKubexFS", logz.QUIET)
	d.StartService("kubex-mongodb", "mongo:latest", []string{fmt.Sprintf("%s:27017", mongoPort)}, nil)
	_ = logz.InfoLog(fmt.Sprintf("✅ MongoDB iniciado com sucesso!"), "GoKubexFS", logz.QUIET)
}
func (d *DatabaseServiceImpl) StartPostgres() {
	postgrePort := viper.GetString("database.port")
	postgreHost := viper.GetString("database.host")
	postgreUser := viper.GetString("database.username")
	postgrePass := viper.GetString("database.password")
	postgreName := viper.GetString("database.name")
	postgrePath := viper.GetString("database.path")
	postgreConnStr := viper.GetString("database.connection_string")

	if d.IsServiceRunning("kubex-postgres") {
		_ = logz.InfoLog(fmt.Sprintf("✅ Postgres já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando Postgres..."), "GoKubexFS", logz.QUIET)
	d.StartService("kubex-postgres", "postgres:latest", []string{fmt.Sprintf("%s:5432", postgrePort)}, []string{
		fmt.Sprintf("POSTGRES_CONNECTION_STRING=%s", postgreConnStr),
		fmt.Sprintf("POSTGRES_PATH=%s", postgrePath),
		fmt.Sprintf("POSTGRES_HOST=%s", postgreHost),
		fmt.Sprintf("POSTGRES_USER=%s", postgreUser),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", postgrePass),
		fmt.Sprintf("POSTGRES_DB=%s", postgreName),
		fmt.Sprintf("POSTGRESQL_ENABLE_TLS=%s", "false"),
		fmt.Sprintf("POSTGRESQL_MAX_CONNECTIONS=%s", "100"),
		fmt.Sprintf("POSTGRESQL_VOLUME_DIR=%s", "/var/lib/postgresql/data"),
		fmt.Sprintf("POSTGRESQL_TIMEZONE=%s", "UTC"),
	})
	_ = logz.InfoLog(fmt.Sprintf("✅ Postgres iniciado com sucesso!"), "GoKubexFS", logz.QUIET)

}
func (d *DatabaseServiceImpl) StartRabbitMQ() {
	rabbitPort := viper.GetString("rabbitmq.port")
	rabbitMgmtPort := viper.GetString("rabbitmq.management_port")
	rabbitUser := viper.GetString("rabbitmq.username")
	rabbitPass := viper.GetString("rabbitmq.password")

	if d.IsServiceRunning("kubex-rabbitmq") {
		_ = logz.InfoLog(fmt.Sprintf("✅ RabbitMQ já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando RabbitMQ..."), "GoKubexFS", logz.QUIET)
	d.StartService("kubex-rabbitmq", "rabbitmq:management", []string{
		fmt.Sprintf("%s:5672", rabbitPort),
		fmt.Sprintf("%s:15672", rabbitMgmtPort),
	}, []string{
		fmt.Sprintf("RABBITMQ_DEFAULT_USER=%s", rabbitUser),
		fmt.Sprintf("RABBITMQ_DEFAULT_PASS=%s", rabbitPass),
	})
	_ = logz.InfoLog(fmt.Sprintf("✅ RabbitMQ iniciado com sucesso!"), "GoKubexFS", logz.QUIET)
}
func (d *DatabaseServiceImpl) StartRedis() {
	redisPort := viper.GetString("redis.port")

	if d.IsServiceRunning("kubex-redis") {
		_ = logz.InfoLog(fmt.Sprintf("✅ Redis já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando Redis..."), "GoKubexFS", logz.QUIET)
	d.StartService("kubex-redis", "redis:latest", []string{fmt.Sprintf("%s:6379", redisPort)}, nil)
	_ = logz.InfoLog(fmt.Sprintf("✅ Redis iniciado com sucesso!"), "GoKubexFS", logz.QUIET)
}
func (d *DatabaseServiceImpl) StartService(serviceName, image string, ports []string, envVars []string) {
	if d.IsServiceRunning(serviceName) {
		fmt.Printf("✅ %s já está rodando!\n", serviceName)
		return
	}

	fmt.Printf("🚀 Iniciando %s...\n", serviceName)
	args := []string{"run", "-d", "--name", serviceName}
	for _, port := range ports {
		args = append(args, "-p", port)
	}
	for _, env := range envVars {
		args = append(args, "-e", env)
	}
	args = append(args, image)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Erro ao iniciar %s: %v", serviceName, err)
	}
	fmt.Printf("✅ %s iniciado com sucesso!\n", serviceName)
}
func (d *DatabaseServiceImpl) SetupDatabaseServices() error {
	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando serviços..."), "GoKubexFS", logz.QUIET)
	if existsConfigFileErr := d.fs.ExistsConfigFile(); !existsConfigFileErr {
		cfg := NewConfigService(d.fs.GetConfigFilePath(), d.fs.GetDefaultKeyPath(), d.fs.GetDefaultCertPath())
		if cfgSetupErr := cfg.SetupConfig(); cfgSetupErr != nil {
			return fmt.Errorf(fmt.Sprintf("❌ Erro ao configurar o arquivo de configuração: %v", cfgSetupErr))
		}
	} else {
		viper.SetConfigFile(d.fs.GetConfigFilePath())
		if readConfigErr := viper.ReadInConfig(); readConfigErr != nil {
			return fmt.Errorf(fmt.Sprintf("❌ Erro ao ler o arquivo de configuração: %v", readConfigErr))
		}
	}

	if !d.IsDockerRunning() {
		return fmt.Errorf(fmt.Sprintf("❌ Docker não está rodando!"))
	}
	if !d.ExistsContainer("kubex-postgres") {
		d.StartPostgres()
	}
	if !d.ExistsContainer("kubex-mongodb") {
		d.StartMongoDB()
	}
	if !d.ExistsContainer("kubex-redis") {
		d.StartRedis()
	}
	if !d.ExistsContainer("kubex-rabbitmq") {
		d.StartRabbitMQ()
	}

	return nil
}

func (d *DatabaseServiceImpl) ConnectDB() error {
	d.loadViperConfig()
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
		d.loadViperConfig()
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
	d.loadViperConfig()
	_ = logz.InfoLog("Reconnecting to database...", "GDBase", logz.QUIET)
	d.dbChanCtl <- "reconnect"
	_ = logz.InfoLog("Database reconnection successful!", "GDBase", logz.QUIET)
	return <-d.dbChanErr
}
func (d *DatabaseServiceImpl) IsConnected() error {
	if d != nil {
		if err := d.checkDatabaseHealth(); err != nil {
			return fmt.Errorf("❌ Erro ao verificar a conexão com o banco de dados: %v", err)
		}
		return nil
	} else {
		mtx := &sync.Mutex{}
		mtx.Lock()
		cfg := NewConfigService(d.fs.GetConfigFilePath(), d.fs.GetDefaultKeyPath(), d.fs.GetDefaultCertPath())
		dbCfg := cfg.GetDatabaseConfig()
		d = &DatabaseServiceImpl{
			fs:        NewFileSystemService(""),
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
	d.loadViperConfig()
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
	return d.db, nil
}
func (d *DatabaseServiceImpl) OpenDB() (*gorm.DB, error) {
	d.loadViperConfig()
	if d.db != nil {
		_ = logz.InfoLog("Database connection already open", "GDBase", logz.QUIET)
		return d.db, nil
	}
	var db *gorm.DB
	var dbErr error
	_ = logz.InfoLog("Opening database connection...", "GDBase", logz.QUIET)
	switch d.dbCfg.Type {
	case "oracle", "oci8", "goracle", "godror":
		db, dbErr = d.connectOracle()
	case "mysql", "mariadb":
		db, dbErr = d.connectMySQL()
	case "postgres", "postgresql":
		db, dbErr = d.connectPostgres()
	case "mssql", "sqlserver":
		db, dbErr = d.connectMSSQL()
	case "sqlite", "sqlite3":
		db, dbErr = d.connectSQLite()
	default:
		_ = logz.WarnLog(fmt.Sprintf("Database type not specified or invalid (%s), falling back to SQLite", d.dbCfg.Type), "GDBase", logz.QUIET)
		db, dbErr = d.connectSQLite()
	}
	if dbErr != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("Failed to connect to database: %v", dbErr), "GDBase", logz.QUIET)
	}

	//_ = logz.InfoLog("Checking database health...", "GDBase", logz.QUIET)
	//healthyDB, errorDb := d.initialHealthCheck()
	//if errorDb != nil {
	//	_ = logz.ErrorLog(fmt.Sprintf("Failed to check database health: %v", errorDb), "GDBase", logz.QUIET)
	//	return nil, errorDb
	//}
	//_ = logz.InfoLog("Database health check successful!", "GDBase", logz.QUIET)
	//db = healthyDB

	_ = logz.InfoLog(fmt.Sprintf("Trying to migrate models to %s database...", d.dbCfg.Type), "GDBase", logz.QUIET)
	if migrateErr := d.db.AutoMigrate(dbModels.ModelListExp...); migrateErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Failed to migrate models to %s database: %v", d.dbCfg.Type, migrateErr), "GDBase", logz.QUIET)
		return nil, migrateErr
	}

	_ = logz.InfoLog(fmt.Sprintf("Successfully migrated models to %s database", d.dbCfg.Type), "GDBase", logz.QUIET)
	return db, nil
}
func (d *DatabaseServiceImpl) connectMySQL() (*gorm.DB, error) {
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
		return d.initialHealthCheck()
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) connectPostgres() (*gorm.DB, error) {
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
			return d.initialHealthCheck()
		}
	}
	_ = logz.InfoLog("Checking database health...", "GDBase", logz.QUIET)
	return d.db, nil
}
func (d *DatabaseServiceImpl) connectSQLite() (*gorm.DB, error) {
	var dbErr error
	d.db, dbErr = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if dbErr != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("Initial error connecting to database: %v", dbErr), "GDBase", logz.QUIET)
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) connectMSSQL() (*gorm.DB, error) {
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
		return d.initialHealthCheck()
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) connectOracle() (*gorm.DB, error) {
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
		return d.initialHealthCheck()
	}
	return d.db, nil
}
func (d *DatabaseServiceImpl) checkDatabaseHealth() error {
	if d != nil {
		d.loadViperConfig()
		if d.db == nil {
			return logz.ErrorLog(fmt.Sprintf("❌ Database connection is nil"), "GDBase", logz.QUIET)
		}
		return d.db.Raw("SELECT 1").Error
	}
	return logz.ErrorLog(fmt.Sprintf("❌ Database connection is nil"), "GDBase", logz.QUIET)
}
func (d *DatabaseServiceImpl) waitForDatabase(timeout time.Duration, maxRetries int) error { //, wg *sync.WaitGroup) error {
	_ = logz.InfoLog("Checking database health...", "GDBase", logz.QUIET)
	retryInterval := timeout
	for i := 0; i < maxRetries; i++ {
		if err := d.checkDatabaseHealth(); err == nil {
			return nil
		} else {
			_ = logz.ErrorLog(fmt.Sprintf("Erro ao verificar a conexão com o banco de dados: %v", err), "GDBase", logz.QUIET)
		}
		time.Sleep(retryInterval)
	}
	return logz.ErrorLog(fmt.Sprintf("Falha na conexão com o banco de dados após %d tentativas", maxRetries), "GDBase", logz.QUIET)
}
func (d *DatabaseServiceImpl) initialHealthCheck() (*gorm.DB, error) {
	_ = logz.InfoLog("Checking database health...", "GDBase", logz.QUIET)
	timeout := 2 * time.Second
	maxRetries := 3
	if waitErr := d.waitForDatabase(timeout, maxRetries); waitErr != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("Failed to check database health: %v", waitErr), "GDBase", logz.QUIET)
	} else {
		_ = logz.InfoLog("Database health check successful!", "GDBase", logz.QUIET)
		return d.db, nil
	}
}
func (d *DatabaseServiceImpl) SetConnection(db *gorm.DB) {
	d.loadViperConfig()
	if db != nil {
		_ = logz.InfoLog(fmt.Sprintf("Setting database connection (%s)...", db.Name()), "GDBase")
		d.db = db
		_ = logz.InfoLog("Database connection set successfully!", "GDBase")
	} else {
		_ = logz.WarnLog("Database connection is nil", "GDBase")
	}
}

func (d *DatabaseServiceImpl) ServiceHandler(dbChanData <-chan interface{}) {
	d.loadViperConfig()
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
	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando serviço de banco de dados..."), "GoKubexFS", logz.QUIET)
	cfg := NewConfigService(configFileArg, "", "")
	if stpErr := cfg.SetupConfigFromDbService(); stpErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("❌ Erro ao configurar o arquivo de configuração: %v", stpErr), "GoKubexFS")
		return nil
	}
	dbCfg := cfg.GetDatabaseConfig()

	_ = logz.InfoLog(fmt.Sprintf("🔗 Configurações do banco de dados: %v", dbCfg), "GoKubexFS", logz.QUIET)
	databaseService = &DatabaseServiceImpl{
		fs:        fs,
		db:        nil,
		mtx:       &sync.Mutex{},
		wg:        &sync.WaitGroup{},
		dbCfg:     dbCfg,
		dbChanCtl: make(chan string, 10),
		dbChanErr: make(chan error, 10),
		dbChanSig: make(chan os.Signal, 1),
	}
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
