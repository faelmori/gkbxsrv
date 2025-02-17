package services

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"strings"
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
}

type DatabaseServiceImpl struct {
	fs FileSystemService
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
		log.Fatalf("❌ Erro ao verificar serviço Docker: %v", err)
	}
	return string(output) != ""
}
func (d *DatabaseServiceImpl) ExistsContainer(containerName string) bool {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("❌ Erro ao verificar containers: %v", err)
	}
	return contains(string(output), containerName)
}
func (d *DatabaseServiceImpl) StartMongoDB() {
	mongoPort := viper.GetString("mongodb.port")

	if d.IsServiceRunning("kubex-mongodb") {
		fmt.Println("✅ MongoDB já está rodando!")
		return
	}

	fmt.Println("🚀 Iniciando MongoDB...")
	d.StartService("kubex-mongodb", "mongo:latest", []string{fmt.Sprintf("%s:27017", mongoPort)}, nil)
	fmt.Println("✅ MongoDB iniciado com sucesso!")
}
func (d *DatabaseServiceImpl) StartPostgres() {
	postgrePort := viper.GetString("database.port")
	postgreHost := viper.GetString("database.host")
	postgreUser := viper.GetString("database.user")
	postgrePass := viper.GetString("database.password")
	postgreName := viper.GetString("database.name")
	postgrePath := viper.GetString("database.path")
	postgreConnStr := viper.GetString("database.connection_string")

	if d.IsServiceRunning("kubex-postgres") {
		fmt.Println("✅ Postgres já está rodando!")
		return
	}

	fmt.Println("🚀 Iniciando Postgres...")
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
	fmt.Println("✅ Postgres iniciado com sucesso!")

}
func (d *DatabaseServiceImpl) StartRabbitMQ() {
	rabbitPort := viper.GetString("rabbitmq.port")
	rabbitMgmtPort := viper.GetString("rabbitmq.management_port")
	rabbitUser := viper.GetString("rabbitmq.user")
	rabbitPass := viper.GetString("rabbitmq.password")

	if d.IsServiceRunning("kubex-rabbitmq") {
		fmt.Println("✅ RabbitMQ já está rodando!")
		return
	}

	fmt.Println("🚀 Iniciando RabbitMQ...")
	d.StartService("kubex-rabbitmq", "rabbitmq:management", []string{
		fmt.Sprintf("%s:5672", rabbitPort),
		fmt.Sprintf("%s:15672", rabbitMgmtPort),
	}, []string{
		fmt.Sprintf("RABBITMQ_DEFAULT_USER=%s", rabbitUser),
		fmt.Sprintf("RABBITMQ_DEFAULT_PASS=%s", rabbitPass),
	})
	fmt.Println("✅ RabbitMQ iniciado com sucesso!")
}
func (d *DatabaseServiceImpl) StartRedis() {
	redisPort := viper.GetString("redis.port")

	if d.IsServiceRunning("kubex-redis") {
		fmt.Println("✅ Redis já está rodando!")
		return
	}

	fmt.Println("🚀 Iniciando Redis...")
	d.StartService("kubex-redis", "redis:latest", []string{fmt.Sprintf("%s:6379", redisPort)}, nil)
	fmt.Println("✅ Redis iniciado com sucesso!")
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
	fmt.Println("🚀 Iniciando serviços...")
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
