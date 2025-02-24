package services

import (
	"fmt"
	glb "github.com/faelmori/gokubexfs/internal/globals"
	"github.com/faelmori/kbx/mods/logz"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var fs FileSystemService

type IConfigService interface {
	GetConfigPath() string
	GetSettings() (map[string]interface{}, error)
	GetSetting(key string) (interface{}, error)
	GetLogger() *log.Logger
	GetDatabaseConfig() glb.Database

	SetLogger()

	IsConfigWatchEnabled() bool
	IsConfigLoaded() bool

	WatchConfig(enable bool, event func(fsnotify.Event)) error
	SaveConfig() error
	ResetConfig() error
	LoadConfig() error
	SetupConfig() error

	calculateMD5Hash(filePath string) (string, error)
	getExistingMD5Hash() (string, error)
	saveMD5Hash() error
	compareMD5Hash() (bool, error)
	genCacheFlag(flagToMark string) error
	SetupConfigFromDbService() error
}
type IDockerSrv interface {
	LoadViperConfig() IConfigService
	InstallDocker() error
	IsDockerInstalled() bool
	IsDockerRunning() bool
	IsServiceRunning(serviceName string) bool
	ExistsContainer(containerName string) bool
	StartMongoDB()
	StartPostgres()
	StartRabbitMQ()
	StartRedis()
	StartService(serviceName, image string, ports []string, envVars []string)
	SetupDatabaseServices() error
}
type DockerSrvImpl struct {
	config *IConfigService
}

func (d *DockerSrvImpl) LoadViperConfig() IConfigService {
	cfg := NewConfigSrv("", "", "")
	loadErr := cfg.LoadConfig()
	if loadErr != nil {
		log.Fatalf("❌ Erro ao carregar o arquivo de configuração: %v", loadErr)
	}
	icfg := cfg.(IConfigService)
	return icfg
}
func (d *DockerSrvImpl) IsDockerInstalled() bool {
	cmd := exec.Command("docker", "--version")
	err := cmd.Run()
	return err == nil
}
func (d *DockerSrvImpl) InstallDocker() error {
	if d.IsDockerInstalled() {
		fmt.Println("✅ Docker já está instalado!")
		return nil
	}
	fmt.Println("🚀 Instalando Docker...")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("sh", "-c", "curl -fsSL https://get.docker.com | bash")
	case "darwin":
		cmd = exec.Command("sh", "-c", "brew install --cask docker")
	default:
		return fmt.Errorf("❌ Sistema não suportado")
	}
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}
func (d *DockerSrvImpl) IsDockerRunning() bool {
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Docker não está rodando: %v", err)
	}
	return true
}
func (d *DockerSrvImpl) IsServiceRunning(serviceName string) bool {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", serviceName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		_ = logz.InfoLog(fmt.Sprintf("❌ Erro ao verificar serviço Docker: %v", err), "GoKubexFS", logz.QUIET)
	}
	return string(output) != ""
}
func (d *DockerSrvImpl) ExistsContainer(containerName string) bool {
	d.LoadViperConfig()
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		_ = logz.InfoLog(fmt.Sprintf("❌ Erro ao verificar containers: %v", err), "GoKubexFS", logz.QUIET)
	}
	return contains(string(output), containerName)
}
func (d *DockerSrvImpl) StartMongoDB() {
	mdgConfig := viper.GetStringMapString("mongodb")

	if d.IsServiceRunning("kubex-mongodb") {
		_ = logz.InfoLog(fmt.Sprintf("✅ MongoDB já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando MongoDB..."), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("Info MongoDB:"), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("MONGODB_PORT=%d", mdgConfig["port"]), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("MONGODB_HOST=%s", mdgConfig["host"]), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("MONGODB_USERNAME=%s", mdgConfig["username"]), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("MONGODB_PASSWORD=%s", mdgConfig["password"]), "GoKubexFS", logz.QUIET)
	//_ = logz.DebugLog(fmt.Sprintf("MONGODB_DB=%s", d.config.MongoDB.Database), "GoKubexFS", logz.QUIET)

	d.StartService("kubex-mongodb", "mongo:latest", []string{fmt.Sprintf("%s:27017", mdgConfig["port"])}, nil)
	_ = logz.InfoLog(fmt.Sprintf("✅ MongoDB iniciado com sucesso!"), "GoKubexFS", logz.QUIET)
}
func (d *DockerSrvImpl) StartPostgres() {
	pg := viper.GetStringMapString("database")

	postgrePort := pg["port"]
	postgreHost := pg["host"]
	postgreUser := pg["username"]
	postgrePass := pg["password"]
	postgreName := pg["name"]
	postgrePath := pg["path"]
	postgreConnStr := pg["connection_string"]
	if postgreConnStr == "" {
		postgreConnStr = pg["dsn"]
	}

	if d.IsServiceRunning("kubex-postgres") {
		_ = logz.InfoLog(fmt.Sprintf("✅ Postgres já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando Postgres..."), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("Info Postgres:"), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("POSTGRES_CONNECTION_STRING=%s", postgreConnStr), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("POSTGRES_PORT=%d", postgrePort), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("POSTGRES_PATH=%s", postgrePath), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("POSTGRES_HOST=%s", postgreHost), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("POSTGRES_USERNAME=%s", postgreUser), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("POSTGRES_PASSWORD=%s", postgrePass), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("POSTGRES_DB=%s", postgreName), "GoKubexFS", logz.QUIET)

	d.StartService("kubex-postgres", "postgres:latest", []string{fmt.Sprintf("%s:5432", postgrePort)}, []string{
		fmt.Sprintf("POSTGRES_PORT=%d", postgrePort),
		fmt.Sprintf("POSTGRES_CONNECTION_STRING=%v", postgreConnStr),
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

	// TODO: Trigger dbChanData telling that the database is ready to connect an healthy.

	_ = logz.InfoLog(fmt.Sprintf("✅ Postgres iniciado com sucesso!"), "GoKubexFS", logz.QUIET)

}
func (d *DockerSrvImpl) StartRabbitMQ() {
	rbmqConfig := viper.GetStringMapString("rabbitmq")

	rabbitPort := rbmqConfig["port"]
	rabbitMgmtPort := rbmqConfig["managementPort"]
	rabbitUser := rbmqConfig["username"]
	rabbitPass := rbmqConfig["password"]

	if d.IsServiceRunning("kubex-rabbitmq") {
		_ = logz.InfoLog(fmt.Sprintf("✅ RabbitMQ já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando RabbitMQ..."), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("Info RabbitMQ:"), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("RABBITMQ_PORT=%d", rabbitPort), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("RABBITMQ_MANAGEMENT_PORT=%d", rabbitMgmtPort), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("RABBITMQ_DEFAULT_USER=%s", rabbitUser), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("RABBITMQ_DEFAULT_PASS=%s", rabbitPass), "GoKubexFS", logz.QUIET)

	d.StartService("kubex-rabbitmq", "rabbitmq:management", []string{
		fmt.Sprintf("%s:5672", rabbitPort),
		fmt.Sprintf("%s:15672", rabbitMgmtPort),
	}, []string{
		fmt.Sprintf("RABBITMQ_DEFAULT_USER=%s", rabbitUser),
		fmt.Sprintf("RABBITMQ_DEFAULT_PASS=%s", rabbitPass),
	})
	_ = logz.InfoLog(fmt.Sprintf("✅ RabbitMQ iniciado com sucesso!"), "GoKubexFS", logz.QUIET)
}
func (d *DockerSrvImpl) StartRedis() {
	rdConfig := viper.GetStringMapString("redis")

	redisPort := rdConfig["port"]
	redisUsername := rdConfig["username"]
	redisPassword := rdConfig["password"]
	redisDB := rdConfig["db"]

	if d.IsServiceRunning("kubex-redis") {
		_ = logz.InfoLog(fmt.Sprintf("✅ Redis já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando Redis (%d)...", redisPort), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("Info Redis:"), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("REDIS_PORT=%d", redisPort), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("REDIS_PASSWORD=%s", redisPassword), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("REDIS_USERNAME=%s", redisUsername), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("REDIS_DB=%d", redisDB), "GoKubexFS", logz.QUIET)

	d.StartService("kubex-redis", "redis:latest", []string{
		fmt.Sprintf("%s:6379", redisPort),
	}, []string{
		fmt.Sprintf("REDIS_PORT=%d", redisPort),
		fmt.Sprintf("REDIS_PASSWORD=%s", redisPassword),
		fmt.Sprintf("REDIS_USERNAME=%s", redisUsername),
		fmt.Sprintf("REDIS_DB=%d", redisDB),
	})
	_ = logz.InfoLog(fmt.Sprintf("✅ Redis iniciado com sucesso!"), "GoKubexFS", logz.QUIET)
}
func (d *DockerSrvImpl) StartService(serviceName, image string, ports []string, envVars []string) {
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
		_ = logz.ErrorLog(fmt.Sprintf("❌ Erro ao iniciar %s: %v => %s %s %s", serviceName, err, args, ports, envVars), "GoKubexFS", logz.QUIET)
		os.Exit(1)
	}
	fmt.Printf("✅ %s iniciado com sucesso!\n", serviceName)
}
func (d *DockerSrvImpl) SetupDatabaseServices() error {
	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando serviços..."), "GoKubexFS", logz.QUIET)
	if !d.IsDockerRunning() {
		return logz.ErrorLog(fmt.Sprintf("❌ Docker não está rodando!"), "GoKubexFS")
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

func NewDockerSrv() IDockerSrv {
	return &DockerSrvImpl{}
}

func init() {
	fs = NewFileSystemSrv("")
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
