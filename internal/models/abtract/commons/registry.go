package commons

import (
	"reflect"
)

type ModelRegistry interface {
	GetType() (reflect.Type, error)
	GetCommand() string
	FromModel(model interface{}) ModelRegistry
	FromSerialized(data []byte) (ModelRegistry, error)
	ToModel() Model
}

type RepositoryRegistry interface {
	GetType() (reflect.Type, error)
	FromRepository(repo Repository) RepositoryRegistry
	FromSerialized(data []byte) (RepositoryRegistry, error)
	GetModel() Model
	GetModelType() reflect.Type
	GetRepository() Repository
	GetRepositoryType() reflect.Type
}
