package abstract

import (
	"github.com/faelmori/gkbxsrv/internal/models/abtract/users"
	"reflect"
	"strings"
)

var (
	RepoList = []interface{}{
		&users.RepoTypeUser{},
	}
	ModelRegistryMap = map[string]reflect.Type{
		strings.ToLower("User"): reflect.TypeOf(users.TypeUser{}),
		strings.ToLower("Ping"): reflect.TypeOf(users.TypeUser{}),
	}
	RepoRegistryMap = map[string]reflect.Type{
		strings.ToLower("User"): reflect.TypeOf(users.RepoTypeUser{}),
	}
	RepoRegistryCmdMap = map[string][]string{
		strings.ToLower("User"): []string{"Create", "FindOne", "FindAll", "Update", "Delete", "Close", "List"},
	}
)
