package services

import (
	dkrs "github.com/faelmori/gkbxsrv/internal/services"
)

type DockerSrv = dkrs.IDockerSrv

func NewDockerService() DockerSrv {
	return dkrs.NewDockerSrv()
}
