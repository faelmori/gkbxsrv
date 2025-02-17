package services

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

type DockerService interface {
	IsDockerInstalled() bool
	InstallDocker() error
}
type DockerServiceImpl struct{}

func (d *DockerServiceImpl) IsDockerInstalled() bool {
	cmd := exec.Command("docker", "--version")
	err := cmd.Run()
	return err == nil
}
func (d *DockerServiceImpl) InstallDocker() error {
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

func NewDockerService() DockerService {
	return &DockerServiceImpl{}
}
