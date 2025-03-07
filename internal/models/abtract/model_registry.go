package abstract

import (
	"fmt"
	. "github.com/faelmori/gkbxsrv/internal/models/abtract/commons"
	"github.com/goccy/go-json"
	"reflect"
	"strings"
)

func RegisterModel(name string, modelType reflect.Type) error {
	if _, exists := ModelRegistryMap[strings.ToLower(name)]; exists {
		return fmt.Errorf("model %s já registrado", name)
	}
	ModelRegistryMap[strings.ToLower(name)] = modelType
	return nil
}

type ModelRegistryImpl struct {
	Tp  string `json:"type"`
	Cmd string `json:"command,default=findAll"`
	Dt  Model  `json:"data"`
}

func (m *ModelRegistryImpl) GetType() (reflect.Type, error) {
	if tp, ok := ModelRegistryMap[m.Tp]; ok {
		return tp, nil
	} else {
		return nil, fmt.Errorf("model %s not found", m.Tp)
	}
}
func (m *ModelRegistryImpl) GetCommand() string { return m.Cmd }
func (m *ModelRegistryImpl) FromModel(model interface{}) ModelRegistry {
	m.Tp = strings.ToLower(reflect.TypeOf(model).Name())
	m.Dt = model.(Model)
	return m
}
func (m *ModelRegistryImpl) FromSerialized(data []byte) (ModelRegistry, error) {
	var mdr ModelRegistryImpl
	if err := json.Unmarshal(data, &mdr); err != nil {
		return nil, err
	}
	if _, ok := ModelRegistryMap[mdr.Tp]; !ok {
		return nil, fmt.Errorf("model %s not found", mdr.Tp)
	}
	return &mdr, nil
}
func (m *ModelRegistryImpl) ToModel() Model {
	if tp, ok := ModelRegistryMap[m.Tp]; ok {
		instance := reflect.New(tp).Interface()
		// Preenche os dados se forem deserializáveis
		if m.Dt != nil {
			dataBytes, _ := json.Marshal(m.Dt)
			_ = json.Unmarshal(dataBytes, &instance)
		}
		return instance.(Model)
	}
	return nil
}

func NewModelRegistry() ModelRegistry {
	return &ModelRegistryImpl{}
}
func NewModelRegistryFromModel(model interface{}) ModelRegistry {
	mr := ModelRegistryImpl{}
	return mr.FromModel(model)
}
func NewModelRegistryFromSerialized(data []byte) (ModelRegistry, error) {
	mr := ModelRegistryImpl{}
	return mr.FromSerialized(data)
}
