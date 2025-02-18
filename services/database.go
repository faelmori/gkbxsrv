package services

import (
	"fmt"
	"github.com/faelmori/gokubexfs/internal/globals"
	"github.com/faelmori/kbx/mods/logz"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

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
	ConnectDB() error
	GetDB() (*gorm.DB, error)
	CloseDBConnection() error
	ServiceHandler(dbChanData <-chan interface{})
	Reconnect() error
	IsConnected() error
	GetHost() (string, error)
	GetConnection(client string) (*gorm.DB, error)
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
