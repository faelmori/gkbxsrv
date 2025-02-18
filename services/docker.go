package services

import (
	"fmt"
	"github.com/faelmori/kbx/mods/logz"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"runtime"
)

type DockerSrv interface {
	LoadViperConfig() ConfigServiceImpl
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
	config *ConfigServiceImpl
}

func (d *DockerSrvImpl) LoadViperConfig() ConfigServiceImpl {
	cfg := NewConfigService(Fs.GetConfigFilePath(), Fs.GetDefaultKeyPath(), Fs.GetDefaultCertPath())

	cfgLoadErr := cfg.LoadConfig()
	if cfgLoadErr != nil {
		log.Fatalf("❌ Erro ao carregar configuração: %v", cfgLoadErr)
	}

	vpDBCfg := viper.GetStringMapString("database")
	dbCfg := &Database{
		Type:             vpDBCfg["type"],
		Driver:           vpDBCfg["driver"],
		ConnectionString: vpDBCfg["connection_string"],
		Dsn:              vpDBCfg["dsn"],
		Path:             vpDBCfg["path"],
		Host:             vpDBCfg["host"],
		Port:             vpDBCfg["port"],
		Username:         vpDBCfg["username"],
		Password:         vpDBCfg["password"],
		Name:             vpDBCfg["name"],
	}
	vpRdsCfg := viper.GetStringMapString("redis")
	rdsCfg := &Redis{
		Enabled:  viper.GetBool("redis.enabled"),
		Addr:     vpRdsCfg["host"],
		Port:     vpRdsCfg["port"],
		Username: vpRdsCfg["username"],
		Password: vpRdsCfg["password"],
		DB:       vpRdsCfg["db"],
	}
	vpRbtCfg := viper.GetStringMapString("rabbitmq")
	rbtCfg := &RabbitMQ{
		Enabled:        viper.GetBool("rabbitmq.enabled"),
		Port:           vpRbtCfg["port"],
		ManagementPort: vpRbtCfg["management_port"],
		Username:       vpRbtCfg["username"],
		Password:       vpRbtCfg["password"],
	}
	vpMngCfg := viper.GetStringMapString("mongodb")
	mngCfg := &MongoDB{
		Enabled:  viper.GetBool("mongodb.enabled"),
		Host:     vpMngCfg["host"],
		Port:     vpMngCfg["port"],
		Username: vpMngCfg["username"],
		Password: vpMngCfg["password"],
	}
	config := ConfigServiceImpl{
		Database: *dbCfg,
		Redis:    *rdsCfg,
		RabbitMQ: *rbtCfg,
		MongoDB:  *mngCfg,
	}

	d.config = &config

	return config
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
	fullConfig := d.LoadViperConfig()

	if d.IsServiceRunning("kubex-mongodb") {
		_ = logz.InfoLog(fmt.Sprintf("✅ MongoDB já está rodando!"), "GoKubexFS", logz.QUIET)
		return
	}

	_ = logz.InfoLog(fmt.Sprintf("🚀 Iniciando MongoDB..."), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("Info MongoDB:"), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("MONGODB_PORT=%d", fullConfig.MongoDB.Port), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("MONGODB_HOST=%s", fullConfig.MongoDB.Host), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("MONGODB_USERNAME=%s", fullConfig.MongoDB.Username), "GoKubexFS", logz.QUIET)
	_ = logz.DebugLog(fmt.Sprintf("MONGODB_PASSWORD=%s", fullConfig.MongoDB.Password), "GoKubexFS", logz.QUIET)
	//_ = logz.DebugLog(fmt.Sprintf("MONGODB_DB=%s", d.config.MongoDB.Database), "GoKubexFS", logz.QUIET)

	d.StartService("kubex-mongodb", "mongo:latest", []string{fmt.Sprintf("%s:27017", fullConfig.MongoDB.Port)}, nil)
	_ = logz.InfoLog(fmt.Sprintf("✅ MongoDB iniciado com sucesso!"), "GoKubexFS", logz.QUIET)
}
func (d *DockerSrvImpl) StartPostgres() {
	fullConfig := d.LoadViperConfig()
	postgrePort := fullConfig.Database.Port
	postgreHost := fullConfig.Database.Host
	postgreUser := fullConfig.Database.Username
	postgrePass := fullConfig.Database.Password
	postgreName := fullConfig.Database.Name
	postgrePath := fullConfig.Database.Path
	postgreConnStr := fullConfig.Database.ConnectionString
	if postgreConnStr == "" {
		postgreConnStr = fullConfig.Database.Dsn
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
	d.LoadViperConfig()
	rabbitPort := d.config.RabbitMQ.Port
	rabbitMgmtPort := d.config.RabbitMQ.ManagementPort
	rabbitUser := d.config.RabbitMQ.Username
	rabbitPass := d.config.RabbitMQ.Password

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
	d.LoadViperConfig()
	redisPort := d.config.Redis.Port
	redisUsername := d.config.Redis.Username
	redisPassword := d.config.Redis.Password
	redisDB := d.config.Redis.DB

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

	_ = logz.InfoLog(fmt.Sprintf("🚀 Serviços iniciados com sucesso:"), "GoKubexFS")
	_ = logz.InfoLog(fmt.Sprintf("✅ Postgres: %s", d.config.Database.Dsn), "GoKubexFS")
	_ = logz.InfoLog(fmt.Sprintf("✅ MongoDB: mongodb://%s:%d", d.config.MongoDB.Host, d.config.MongoDB.Port), "GoKubexFS")
	_ = logz.InfoLog(fmt.Sprintf("✅ Redis: redis://%s:%d", d.config.Redis.Addr, d.config.Redis.Port), "GoKubexFS")
	_ = logz.InfoLog(fmt.Sprintf("✅ RabbitMQ: amqp://%s:%d", d.config.RabbitMQ.ManagementPort, d.config.RabbitMQ.Port), "GoKubexFS")

	return nil
}

func NewDockerSrv() DockerSrv {
	return &DockerSrvImpl{}
}
