package services

import (
	"database/sql"
	"fmt"
	"github.com/faelmori/gdbase/models"
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

var ModelList = []interface{}{
	&models.User{},
	&models.Product{},
	&models.Customer{},
	&models.Order{},
}

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
}

type DatabaseServiceImpl struct {
	fs        FileSystemService
	db        *gorm.DB
	mux       *sync.Mutex
	wg        *sync.WaitGroup
	dbCfg     Database
	dbChanCtl chan string
	dbChanErr chan error
	dbChanSig chan os.Signal
	mdRepo    *globals.GenericRepo
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
	dsn := viper.GetString("database.connection_string")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", err)
	}
	d.db = db
	return nil
}
func (d *DatabaseServiceImpl) GetDB() (*gorm.DB, error) {
	if d.db == nil {
		return nil, logz.ErrorLog(fmt.Sprintf("❌ Banco de dados não conectado"), "GoKubexFS")
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
func (d *DatabaseServiceImpl) ServiceHandler(dbChanData <-chan interface{}) {
	for {
		select {
		case dbCtl := <-d.dbChanCtl:
			d.mux.Lock()
			switch dbCtl {
			case "reconnect":
				db := *d.db
				dbClientB, dbClientBErr := db.DB()
				if dbClientBErr != nil {
					d.dbChanErr <- &globals.ValidationError{Message: dbClientBErr.Error(), Field: "dbChanCtl_reconnect"}
				} else {
					if dbClientBErrB := dbClientB.Close(); dbClientBErrB != nil {
						d.dbChanErr <- &globals.ValidationError{Message: dbClientBErrB.Error(), Field: "dbChanCtl_reconnect"}
					} else {
						d.dbChanErr <- nil
					}
				}
			case "isConnected":
				if _, dbErr := d.db.DB(); dbErr != nil {
					d.dbChanErr <- &globals.ValidationError{Message: dbErr.Error(), Field: "dbChanCtl_isConnected"}
				} else {
					d.dbChanErr <- nil
				}
			case "getHost":
				dbCfgB := &d.dbCfg
				d.dbChanErr <- &globals.ValidationError{Message: dbCfgB.Host, Field: "dbChanCtl_getHost"}
			case "close":
				dbB, dbBErr := d.db.DB()
				if dbBErr != nil {
					d.dbChanErr <- &globals.ValidationError{Message: dbBErr.Error(), Field: "dbChanCtl_close"}
				} else {
					_ = dbB.Close()
					d.dbChanErr <- nil
				}
			}
			d.mux.Unlock()
		}
	}
}
func (d *DatabaseServiceImpl) Reconnect() error {
	d.dbChanCtl <- "reconnect"
	return <-d.dbChanErr
}
func (d *DatabaseServiceImpl) IsConnected() error {
	d.dbChanCtl <- "isConnected"
	return <-d.dbChanErr
}
func (d *DatabaseServiceImpl) GetDBConfig(name string) (Database, error) {
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
	return d.db, nil
}
func (d *DatabaseServiceImpl) OpenDB() (*gorm.DB, error) {
	if d.db != nil {
		return d.db, nil
	}
	var db *gorm.DB
	var dbErr error
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

	_ = logz.InfoLog("Checking database health...", "GDBase", logz.QUIET)
	healthyDB, errorDb := d.initialHealthCheck()
	if errorDb != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Failed to check database health: %v", errorDb), "GDBase", logz.QUIET)
		return nil, errorDb
	}
	_ = logz.InfoLog("Database health check successful!", "GDBase", logz.QUIET)
	db = healthyDB

	_ = logz.InfoLog("Getting database connection...", "GDBase", logz.QUIET)
	sqlDb, sqlDbErr := db.DB()
	if sqlDbErr != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("Failed to get database connection: %v", sqlDbErr), "GDBase", logz.QUIET)
	}

	pingAErr := db.ConnPool.(*sql.DB).Ping()
	if pingAErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Failed on first ping to database: %v", pingAErr), "GDBase", logz.QUIET)
		return nil, pingAErr
	}
	_ = logz.InfoLog(fmt.Sprintf("Connected to %s database", d.dbCfg.Type), "GDBase", logz.QUIET)

	_ = logz.InfoLog("Setting database connection pool...", "GDBase", logz.QUIET)
	sqlDb.SetConnMaxIdleTime(time.Minute * 10)
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(100)
	sqlDb.SetConnMaxLifetime(time.Hour)

	_ = logz.InfoLog(fmt.Sprintf("Trying to migrate models to %s database...", d.dbCfg.Type), "GDBase", logz.QUIET)
	if migrateErr := d.db.AutoMigrate(ModelList...); migrateErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Failed to migrate models to %s database: %v", d.dbCfg.Type, migrateErr), "GDBase", logz.QUIET)
		return nil, migrateErr
	}

	pingErr := sqlDb.Ping()
	if pingErr != nil {
		_ = logz.ErrorLog(fmt.Sprintf("Failed to ping database: %v", pingErr), "GDBase", logz.QUIET)
		return nil, pingErr
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
			return d.initialHealthCheck()
		}
	}
	return d.initialHealthCheck()
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
	return d.db.Raw("SELECT 1").Error
}
func (d *DatabaseServiceImpl) waitForDatabase(timeout time.Duration, maxRetries int) error { //, wg *sync.WaitGroup) error {
	retryInterval := timeout
	for i := 0; i < maxRetries; i++ {
		err := d.checkDatabaseHealth()
		if err == nil {
			_ = logz.InfoLog("Conexão com o banco de dados bem-sucedida!", "GDBase", logz.QUIET)
			return nil // Conexão bem-sucedida
		}
		_ = logz.WarnLog(fmt.Sprintf("Falha na conexão com o banco de dados: %v", err), "GDBase", logz.QUIET)
		time.Sleep(retryInterval)
	}
	return logz.ErrorLog(fmt.Sprintf("Falha na conexão com o banco de dados após %d tentativas", maxRetries), "GDBase", logz.QUIET)
}
func (d *DatabaseServiceImpl) initialHealthCheck() (*gorm.DB, error) {
	timeout := 3 * time.Second
	maxRetries := 3
	if waitErr := d.waitForDatabase(timeout, maxRetries); waitErr != nil {
		return nil, waitErr
	} else {
		return d.db, nil
	}
}

func NewDatabaseService(configFileArg string) DatabaseServiceImpl {
	return DatabaseServiceImpl{fs: NewFileSystemService(configFileArg)}
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
