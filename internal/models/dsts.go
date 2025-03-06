package models

import (
	"fmt"
	"github.com/goccy/go-json"
	"reflect"
	"strings"
)

var ModelList = []interface{}{
	&UserImpl{},
	&productImpl{},
	&CustomerImpl{},
	&OrderImpl{},
}
var ModelRegistryMap = map[string]reflect.Type{
	strings.ToLower("User"):     reflect.TypeOf(UserImpl{}),
	strings.ToLower("Product"):  reflect.TypeOf(productImpl{}),
	strings.ToLower("Customer"): reflect.TypeOf(CustomerImpl{}),
	strings.ToLower("Order"):    reflect.TypeOf(OrderImpl{}),
	strings.ToLower("Ping"):     reflect.TypeOf(PingImpl{}),
}

type Model interface {
	TableName() string
	GetID() string
	Validate() error
}
type ModelRegistry interface {
	GetType() (reflect.Type, error)
	GetCommand() string
	FromModel(model interface{}) ModelRegistry
	FromSerialized(data []byte) (ModelRegistry, error)
	ToModel() Model
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

func RegisterModel(name string, modelType reflect.Type) error {
	if _, exists := ModelRegistryMap[strings.ToLower(name)]; exists {
		return fmt.Errorf("model %s já registrado", name)
	}
	ModelRegistryMap[strings.ToLower(name)] = modelType
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

var RepoList = []interface{}{
	&UserRepoImpl{},
	&ProductRepoImpl{},
	&CustomerRepoImpl{},
	&OrderRepoImpl{},
}
var RepoRegistryMap = map[string]reflect.Type{
	strings.ToLower("User"):     reflect.TypeOf(UserRepoImpl{}),
	strings.ToLower("Product"):  reflect.TypeOf(ProductRepoImpl{}),
	strings.ToLower("Customer"): reflect.TypeOf(CustomerRepoImpl{}),
	strings.ToLower("Order"):    reflect.TypeOf(OrderRepoImpl{}),
}
var RepoRegistryCmdMap = map[string][]string{
	strings.ToLower("User"):     []string{"Create", "FindOne", "FindAll", "Update", "Delete", "Close", "List"},
	strings.ToLower("Product"):  []string{"Create", "FindOne", "FindAll", "Update", "Delete", "Close", "List", "FindAllByDepart", "FindAllByCategory"},
	strings.ToLower("Customer"): []string{"Create", "FindOne", "FindAll", "Update", "Delete", "Close", "List"},
	strings.ToLower("Order"):    []string{"Create", "FindOne", "FindAll", "Update", "Delete", "Close", "List"},
}

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
type RepositoryRegistry interface {
	GetType() (reflect.Type, error)
	FromRepository(repo Repository) RepositoryRegistry
	FromSerialized(data []byte) (RepositoryRegistry, error)
	GetModel() Model
	GetModelType() reflect.Type
	GetRepository() Repository
	GetRepositoryType() reflect.Type
}
type RepoRegistryImpl struct {
	Tp      string   `json:"type"`
	Md      string   `json:"modelType"`
	CmdList []string `json:"commands"`
}

func (r *RepoRegistryImpl) GetType() (reflect.Type, error) {
	if tp, ok := RepoRegistryMap[r.Tp]; ok {
		return tp, nil
	} else {
		return nil, fmt.Errorf("repositório %s não encontrado", r.Tp)
	}
}
func (r *RepoRegistryImpl) FromRepository(repo Repository) RepositoryRegistry {
	r.Tp = strings.ToLower(reflect.TypeOf(repo).Name())
	r.Md = strings.ToLower(reflect.TypeOf(repo.GetModel()).Name())
	r.CmdList = RepoRegistryCmdMap[r.Md]
	return r
}
func (r *RepoRegistryImpl) FromSerialized(data []byte) (RepositoryRegistry, error) {
	var rr RepoRegistryImpl
	if err := json.Unmarshal(data, &rr); err != nil {
		return nil, err
	}
	if _, ok := RepoRegistryMap[rr.Tp]; !ok {
		return nil, fmt.Errorf("repositório %s não encontrado", rr.Tp)
	}
	return &rr, nil
}
func (r *RepoRegistryImpl) GetModel() Model {
	if tp, ok := ModelRegistryMap[r.Md]; ok {
		instance := reflect.New(tp).Interface()
		return instance.(Model)
	}
	return nil
}
func (r *RepoRegistryImpl) GetModelType() reflect.Type {
	if tp, ok := ModelRegistryMap[r.Md]; ok {
		return tp
	}
	return nil
}
func (r *RepoRegistryImpl) GetRepository() Repository {
	if tp, ok := RepoRegistryMap[r.Tp]; ok {
		instance := reflect.New(tp).Interface()
		return instance.(Repository)
	}
	return nil
}
func (r *RepoRegistryImpl) GetRepositoryType() reflect.Type {
	if tp, ok := RepoRegistryMap[r.Tp]; ok {
		return tp
	}
	return nil
}

func RegisterRepository(name string, repoType reflect.Type) error {
	if _, exists := RepoRegistryMap[strings.ToLower(name)]; exists {
		return fmt.Errorf("repositório %s já registrado", name)
	}
	RepoRegistryMap[strings.ToLower(name)] = repoType
	return nil
}
func NewRepoRegistry() RepositoryRegistry {
	return &RepoRegistryImpl{}
}
func NewRepoRegistryFromRepository(repo Repository) RepositoryRegistry {
	rr := RepoRegistryImpl{}
	return rr.FromRepository(repo)
}
func NewRepoRegistryFromSerialized(data []byte) (RepositoryRegistry, error) {
	rr := RepoRegistryImpl{}
	return rr.FromSerialized(data)
}
