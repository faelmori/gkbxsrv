package models

import (
	"github.com/faelmori/gkbxsrv/internal/models"
	"reflect"
)

var ModelListExp = []interface{}{
	Ping{},
	User{},
	Product{},
	Customer{},
	Order{},
}

type ModelRegistry = models.ModelRegistry

func RegisterModel(name string, modelType reflect.Type) error {
	return models.RegisterModel(name, modelType)
}
func NewModelRegistry() ModelRegistry { return models.NewModelRegistry() }
func NewModelRegistryFromModel(model interface{}) ModelRegistry {
	return models.NewModelRegistryFromModel(model)
}
