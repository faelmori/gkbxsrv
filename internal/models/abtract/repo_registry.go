package abstract

import (
	"fmt"
	. "github.com/faelmori/gkbxsrv/internal/models/abtract/commons"
	"github.com/goccy/go-json"
	"reflect"
	"strings"
)

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
