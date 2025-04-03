package models

import glb "github.com/faelmori/gkbxsrv/internal/globals"

type Server interface{ *glb.Server }
type Database interface{ *glb.Database }
type JWT interface{ *glb.JWT }
type Redis interface{ *glb.Redis }
type RabbitMQ interface{ *glb.RabbitMQ }
type MongoDB interface{ *glb.MongoDB }
type Certificate interface{ *glb.Certificate }
type Docker interface{ *glb.Docker }
type FileSystem interface{ *glb.FileSystem }
type Cache interface{ *glb.Cache }
type ValidationError interface{ *glb.ValidationError }
