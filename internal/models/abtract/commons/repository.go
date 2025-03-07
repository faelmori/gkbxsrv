package commons

import (
	"reflect"
)

type Repository interface {
	GetModel() reflect.Type
	Create(u Model) (Model, error)
	FindOne(where ...interface{}) (Model, error)
	FindAll(where ...interface{}) ([]Model, error)
	Update(u Model) (Model, error)
	Delete(id uint) error
	Close() error
	List(where ...interface{}) (TableHandler, error)
	ExecuteCommand(command string, data interface{}) (interface{}, error)
}
