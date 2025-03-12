package models

import (
	"fmt"
	"github.com/goccy/go-json"
	"reflect"
	"strings"
)

type Model interface {
	Validate() error
}

var ModelList = []interface{}{
	&UserImpl{},
	&Product{},
	&CustomerImpl{},
	&Order{},
}
var ModelRegistryMap = map[string]reflect.Type{
	strings.ToLower("User"):     reflect.TypeOf(UserImpl{}),
	strings.ToLower("Product"):  reflect.TypeOf(Product{}),
	strings.ToLower("Customer"): reflect.TypeOf(CustomerImpl{}),
	strings.ToLower("Order"):    reflect.TypeOf(Order{}),
	strings.ToLower("Ping"):     reflect.TypeOf(PingImpl{}),
}

type ModelRegistryImpl struct {
	Tp string      `json:"type"`
	Dt interface{} `json:"data"`
}
type ModelRegistryInterface interface {
	GetType() (reflect.Type, error)
	FromModel(model interface{}) ModelRegistryInterface
	FromSerialized(data []byte) (ModelRegistryInterface, error)
	ToModel() interface{}
}

func (m *ModelRegistryImpl) GetType() (reflect.Type, error) {
	if tp, ok := ModelRegistryMap[m.Tp]; ok {
		return tp, nil
	} else {
		return nil, fmt.Errorf("model %s not found", m.Tp)
	}
}
func (m *ModelRegistryImpl) FromModel(model interface{}) ModelRegistryInterface {
	m.Tp = strings.ToLower(reflect.TypeOf(model).Name())
	m.Dt = model
	return m
}
func (m *ModelRegistryImpl) FromSerialized(data []byte) (ModelRegistryInterface, error) {
	var mdr ModelRegistryImpl
	if err := json.Unmarshal(data, &mdr); err != nil {
		return nil, err
	}
	if _, ok := ModelRegistryMap[mdr.Tp]; !ok {
		return nil, fmt.Errorf("model %s not found", mdr.Tp)
	}
	return &mdr, nil
}
func (m *ModelRegistryImpl) ToModel() interface{} {
	if tp, ok := ModelRegistryMap[m.Tp]; ok {
		instance := reflect.New(tp).Interface()
		// Preenche os dados se forem deserializáveis
		if m.Dt != nil {
			dataBytes, _ := json.Marshal(m.Dt)
			_ = json.Unmarshal(dataBytes, &instance)
		}
		return instance
	}
	return nil
}

func RegisterModel(name string, modelType reflect.Type) error {
	if _, exists := ModelRegistryMap[strings.ToLower(name)]; exists {
		return fmt.Errorf("model %s já registrado", name)
	}
	ModelRegistryMap[strings.ToLower(name)] = modelType
	return nil
}
func NewModelRegistry() ModelRegistryInterface {
	return &ModelRegistryImpl{}
}
func NewModelRegistryFromModel(model interface{}) ModelRegistryInterface {
	mr := ModelRegistryImpl{}
	return mr.FromModel(model)
}
func NewModelRegistryFromSerialized(data []byte) (ModelRegistryInterface, error) {
	mr := ModelRegistryImpl{}
	return mr.FromSerialized(data)
}
