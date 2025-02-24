package services

import (
	dkrs "github.com/faelmori/gokubexfs/internal/services"
)

type DockerSrv = dkrs.IDockerSrv

func NewDockerService() DockerSrv {
	return dkrs.NewDockerSrv()
}
