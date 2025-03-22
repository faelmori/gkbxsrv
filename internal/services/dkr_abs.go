package services

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"github.com/faelmori/logz"
	"io"
	"strconv"

	c "github.com/docker/docker/api/types/container"
	i "github.com/docker/docker/api/types/image"
	v "github.com/docker/docker/api/types/volume"
	k "github.com/docker/docker/client"
	n "github.com/docker/go-connections/nat"

	"github.com/faelmori/gkbxsrv/internal/utils"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed assets/init-db.sql
var initDBSQL []byte

var fs FileSystemService

type IDockerSrv interface {
	LoadViperConfig() ConfigService
	InstallDocker() error
	IsDockerInstalled() bool
	IsDockerRunning() bool
	IsServiceRunning(serviceName string) bool
	ExistsContainer(containerName string) bool
	StartMongoDB() error
	StartPostgres() error
	StartRabbitMQ() error
	StartRedis() error
	StartService(serviceName string, image string, ports n.PortSet, envVars []string, volumes map[string]struct{}, portBindings n.PortMap) error
	SetupDatabaseServices() error
	GetContainerLogs(containerName string) error
	createVolume(volumeName, devicePath string) error
	writeInitDBSQL() (string, error)
}
type DockerSrvImpl struct {
	config  ConfigService
	volumes map[string]map[string]struct{}
}

func (d *DockerSrvImpl) SetupDatabaseServices() error {
	if !d.IsDockerRunning() {
		return fmt.Errorf("❌ Docker não está rodando")
	}

	if !d.ExistsContainer("gkbxsrv-pg") {
		if pgErr := d.StartPostgres(); pgErr != nil {
			return pgErr
		}
	}

	/*if !d.ExistsContainer("kubex-mongodb") {
		d.StartMongoDB()
	}*/

	if !d.ExistsContainer("kubex-redis") {
		redisErr := d.StartRedis()
		if redisErr != nil {
			return redisErr
		}
	}

	/*if !d.ExistsContainer("kubex-rabbitmq") {
		d.StartRabbitMQ()
	}*/
	return nil
}

func (d *DockerSrvImpl) LoadViperConfig() ConfigService {
	cfg := NewConfigSrv("", "", "")
	loadErr := cfg.LoadConfig()
	if loadErr != nil {
		log.Fatalf("❌ Erro ao carregar o arquivo de configuração: %v", loadErr)
	}
	icfg := cfg.(ConfigService)
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
		fmt.Printf("❌ Erro ao verificar containers: %v\n", err)
	}
	return string(output) != ""
}
func (d *DockerSrvImpl) ExistsContainer(containerName string) bool {
	d.LoadViperConfig()
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Erro ao verificar containers: %v\n", err)
	}
	return contains(string(output), containerName)
}
func (d *DockerSrvImpl) GetContainerLogs(containerName string) error {
	cli, err := k.NewClientWithOpts(k.FromEnv, k.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("erro ao criar cliente do Docker: %w", err)
	}
	ctx := context.Background()
	logsReader, err := cli.ContainerLogs(ctx, containerName, c.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     false,
	})
	if err != nil {
		return fmt.Errorf("erro ao obter logs do container %s: %w", containerName, err)
	}
	defer func(logsReader io.ReadCloser) {
		_ = logsReader.Close()
	}(logsReader)
	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if scannerErr := scanner.Err(); scannerErr != nil {
		return fmt.Errorf("erro ao processar logs do container %s: %w", containerName, scannerErr)
	}
	return nil
}
func (d *DockerSrvImpl) createVolume(volumeName, devicePath string) error {
	cli, err := k.NewClientWithOpts(k.FromEnv, k.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	vol, volErr := cli.VolumeCreate(context.Background(), v.CreateOptions{
		Name:   volumeName,
		Driver: "local",
		DriverOpts: map[string]string{
			"type":   "none",
			"device": devicePath,
			"o":      "bind",
		},
	})

	if volErr != nil {
		return volErr
	}

	fmt.Printf("Volume created: %s\n", vol.Name)
	return nil
}
func (d *DockerSrvImpl) writeInitDBSQL() (string, error) {
	cwdDir, cwdDirErr := utils.GetWorkDir()
	if cwdDirErr != nil {
		return "", cwdDirErr
	}
	configDir := filepath.Join(cwdDir, "databases", "pg", "init")
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return "", err
	}
	filePath := filepath.Join(configDir, "init-db.sql")
	if err := os.WriteFile(filePath, initDBSQL, 0644); err != nil {
		return "", err
	}
	return filePath, nil
}
func (d *DockerSrvImpl) findAvailablePort(basePort int, maxAttempts int) (string, error) {
	for i := 0; i < maxAttempts; i++ {
		port := fmt.Sprintf("%d", basePort+i)
		isOpen, err := utils.CheckPortOpen(port)
		if err != nil {
			return "", fmt.Errorf("erro ao verificar porta %s: %w", port, err)
		}
		if !isOpen {
			fmt.Printf("⚠️ Porta %s está ocupada, tentando a próxima...\n", port)
			continue
		}
		fmt.Printf("✅ Porta disponível encontrada: %s\n", port)
		return port, nil
	}
	return "", fmt.Errorf("nenhuma porta disponível no range %d-%d", basePort, basePort+maxAttempts-1)
}

func (d *DockerSrvImpl) StartPostgres() error {
	pg := viper.GetStringMapString("database")

	if d.IsServiceRunning("gkbxsrv-pg") {
		fmt.Printf("✅ PostgreSQL já está rodando!\n")
		return nil
	}

	basePort, _ := strconv.Atoi(pg["port"])
	port, err := d.findAvailablePort(basePort, 3)
	if err != nil {
		return fmt.Errorf("erro ao encontrar porta para PostgreSQL: %w", err)
	}

	savedPort := pg["port"]
	if savedPort == "" {
		port, err = d.findAvailablePort(basePort, 3)
		if err != nil {
			return fmt.Errorf("erro ao encontrar porta para MongoDB: %w", err)
		}
		dbConfig := d.config.GetDatabaseConfig()
		dbConfig.Port = port

		if err := d.config.SetDatabaseConfig(&dbConfig); err != nil {
			fmt.Printf("⚠️ Erro ao persistir porta para MongoDB: %v\n", err)
		} else {
			fmt.Printf("💾 Porta para MongoDB persistida: %s\n", port)
		}
	}

	prtPort := n.Port(port + "/tcp")
	ports := n.PortSet{
		prtPort: struct{}{},
	}

	portBindings := n.PortMap{
		prtPort: []n.PortBinding{
			{
				HostIP:   pg["host"],
				HostPort: port,
			},
		},
	}

	initVolPath, initVolErr := d.writeInitDBSQL()
	if initVolErr != nil {
		return fmt.Errorf("falha ao criar arquivo de inicialização: %w", initVolErr)
	}
	initVolDir := filepath.Join(filepath.Dir(initVolPath))
	dataVolDir := filepath.Join(filepath.Dir(initVolDir), "data")

	volumes := make(map[string]struct{})
	volumes["gkbxsrv-pg-init"] = struct{}{}
	volumes["gkbxsrv-pg-data"] = struct{}{}

	if err := d.createVolume("gkbxsrv-pg-init", initVolDir); err != nil {
		return fmt.Errorf("falha ao criar volume de init: %w", err)
	}
	if err := d.createVolume("gkbxsrv-pg-data", dataVolDir); err != nil {
		return fmt.Errorf("falha ao criar volume de dados: %w", err)
	}

	envVars := []string{
		"POSTGRES_USER=" + pg["username"],
		"POSTGRES_PASSWORD=" + pg["password"],
		"POSTGRES_DB=" + pg["name"],
	}

	return d.StartService(
		"gkbxsrv-pg",
		"postgres:14-alpine",
		ports,
		envVars,
		volumes,
		portBindings,
	)
}
func (d *DockerSrvImpl) StartMongoDB() error {
	mdgConfig := viper.GetStringMapString("mongodb")
	//portKey := "mongodb.port"

	// Recuperar porta persistida, se disponível
	//savedPort, err := d.config.Get(portKey)
	//if err == nil && savedPort != "" {
	//	fmt.Printf("🔄 Usando porta persistida para MongoDB: %s\n", savedPort)
	//	mdgConfig["port"] = savedPort
	//}

	if d.IsServiceRunning("kubex-mongodb") {
		fmt.Printf("✅ MongoDB já está rodando!\n")
		return nil
	}

	// Encontrar uma porta disponível, se necessário
	//basePort, _ := strconv.Atoi(mdgConfig["port"])
	//port := mdgConfig["port"] // Porta padrão ou persistida
	//if savedPort == "" {
	//	port, err = d.findAvailablePort(basePort, 3)
	//	if err != nil {
	//		return fmt.Errorf("erro ao encontrar porta para MongoDB: %w", err)
	//	}
	//
	//	// Persistir a porta encontrada
	//	d.config.GetDatabaseConfig().Port = port
	//	if err := d.config.Set(portKey, port); err != nil {
	//		fmt.Printf("⚠️ Erro ao persistir porta para MongoDB: %v\n", err)
	//	} else {
	//		fmt.Printf("💾 Porta para MongoDB persistida: %s\n", port)
	//	}
	//}

	// Configuração de Portas
	//mongoPort := n.Port(port + "/tcp")

	mongoPort := n.Port("20001/tcp")
	ports := n.PortSet{
		mongoPort: struct{}{},
	}

	portBindings := n.PortMap{
		mongoPort: []n.PortBinding{
			{
				HostIP:   mdgConfig["host"],
				HostPort: strconv.Itoa(20001),
			},
		},
	}

	// Configuração de Volumes
	dataVolDir := filepath.Join("/path/to/mongo/data")
	configVolDir := filepath.Join("/path/to/mongo/config")

	volumes := make(map[string]struct{})
	volumes["kubex-mongo-data"] = struct{}{}
	volumes["kubex-mongo-config"] = struct{}{}

	if err := d.createVolume("kubex-mongo-data", dataVolDir); err != nil {
		return fmt.Errorf("falha ao criar volume de dados: %w", err)
	}
	if err := d.createVolume("kubex-mongo-config", configVolDir); err != nil {
		return fmt.Errorf("falha ao criar volume de configuração: %w", err)
	}

	// Configuração de Variáveis de Ambiente (se necessário)
	envVars := []string{}

	return d.StartService(
		"kubex-mongodb",
		"mongo:latest",
		ports,
		envVars,
		volumes,
		portBindings,
	)
}
func (d *DockerSrvImpl) StartRabbitMQ() error {
	rbmqConfig := viper.GetStringMapString("rabbitmq")

	rmqPort := n.Port("5672/tcp")
	rmqMgmtPort := n.Port("15672/tcp")
	ports := n.PortSet{
		rmqPort:     struct{}{},
		rmqMgmtPort: struct{}{},
	}

	portBindings := n.PortMap{
		rmqPort: []n.PortBinding{
			{
				HostIP:   rbmqConfig["host"],
				HostPort: rbmqConfig["port"],
			},
		},
		rmqMgmtPort: []n.PortBinding{
			{
				HostIP:   rbmqConfig["host"],
				HostPort: rbmqConfig["managementPort"],
			},
		},
	}

	envVars := []string{
		"RABBITMQ_DEFAULT_USER=" + rbmqConfig["username"],
		"RABBITMQ_DEFAULT_PASS=" + rbmqConfig["password"],
	}

	return d.StartService(
		"kubex-rabbitmq",
		"rabbitmq:management",
		ports,
		envVars,
		nil, // Sem volumes específicos
		portBindings,
	)
}
func (d *DockerSrvImpl) StartRedis() error {
	rdConfig := viper.GetStringMapString("redis")

	redisPort := n.Port("6379/tcp")
	ports := n.PortSet{
		redisPort: struct{}{},
	}

	portBindings := n.PortMap{
		redisPort: []n.PortBinding{
			{
				HostIP:   rdConfig["host"],
				HostPort: rdConfig["port"],
			},
		},
	}

	envVars := []string{
		"REDIS_PASSWORD=" + rdConfig["password"],
		"REDIS_USERNAME=" + rdConfig["username"],
		"REDIS_DB=" + rdConfig["db"],
	}

	return d.StartService(
		"kubex-redis",
		"redis:latest",
		ports,
		envVars,
		nil, // Sem volumes específicos
		portBindings,
	)
}

func (d *DockerSrvImpl) StartService(serviceName string, image string, ports n.PortSet, envVars []string, volumes map[string]struct{}, portBindings n.PortMap) error {
	if d.IsServiceRunning(serviceName) {
		fmt.Printf("✅ %s já está rodando!\n", serviceName)
		return nil
	}

	cli, err := k.NewClientWithOpts(k.FromEnv, k.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("erro ao criar cliente do Docker: %w", err)
	}

	ctx := context.Background()

	logz.Info("Verificando imagem... Aguarde um momento...", map[string]interface{}{
		"context":     "StartService",
		"serviceName": serviceName,
		"image":       image,
	})

	reader, readerErr := cli.ImagePull(ctx, image, i.PullOptions{})
	if readerErr != nil {
		logz.Error("Erro ao fazer pull da imagem", map[string]interface{}{
			"context":     "StartService",
			"serviceName": serviceName,
			"image":       image,
			"error":       readerErr.Error(),
		})
		return readerErr
	}
	defer reader.Close()
	_, _ = io.Copy(io.Discard, reader)

	logz.Info("Iniciando container... Aguarde um momento...", map[string]interface{}{
		"context":     "StartService",
		"serviceName": serviceName,
		"image":       image,
	})

	containerConfig := &c.Config{
		Image:        image,
		Env:          envVars,
		ExposedPorts: ports,
		Volumes:      volumes,
	}

	hostConfig := &c.HostConfig{
		PortBindings: portBindings,
	}

	resp, respErr := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, serviceName)
	if respErr != nil {
		logz.Error("Erro ao criar container", map[string]interface{}{
			"context":     "StartService",
			"serviceName": serviceName,
			"image":       image,
			"error":       respErr.Error(),
		})
		return respErr
	}

	if containerStartErr := cli.ContainerStart(ctx, resp.ID, c.StartOptions{}); containerStartErr != nil {
		logz.Error("Erro ao iniciar container", map[string]interface{}{
			"context":     "StartService",
			"serviceName": serviceName,
			"error":       containerStartErr.Error(),
		})
		return containerStartErr
	}

	logz.Info("Container iniciado com sucesso!", map[string]interface{}{
		"context":     "StartService",
		"serviceName": serviceName,
	})

	return nil
}

func NewDockerSrv() IDockerSrv {
	if fs == nil {
		fs = NewFileSystemSrv("")
	}
	return &DockerSrvImpl{
		config:  NewConfigSrv(fs.GetConfigFilePath(), fs.GetDefaultKeyPath(), fs.GetDefaultCertPath()),
		volumes: make(map[string]map[string]struct{}),
	}
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

func init() {
	fs = NewFileSystemSrv("")
}
