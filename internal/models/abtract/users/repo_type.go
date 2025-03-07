package users

import (
	"fmt"
	. "github.com/faelmori/gkbxsrv/internal/models/abtract/commons"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

type RepoTypeUser struct{ *gorm.DB }

func NewRepoUser(db *gorm.DB) RepoUser {
	usr := RepoTypeUser{db}
	return &usr
}

func (g *RepoTypeUser) GetModel() reflect.Type {
	return reflect.TypeOf(&TypeUser{})
}
func (g *RepoTypeUser) Create(u Model) (Model, error) {
	usr := *u.(*TypeUser)
	err := g.DB.Create(&usr).Error
	if err != nil {
		return nil, fmt.Errorf("TypeUser repository: failed to create TypeUser: %w", err)
	}
	return u, nil
}
func (g *RepoTypeUser) FindOne(where ...interface{}) (Model, error) {
	var u *TypeUser
	err := g.DB.Where(where[0], where[1:]).First(&u).Error
	if err != nil {
		return nil, fmt.Errorf("TypeUser repository: failed to find TypeUser: %w", err)
	}
	return u, nil
}
func (g *RepoTypeUser) FindAll(where ...interface{}) ([]Model, error) {
	var us []TypeUser
	err := g.DB.Where(where[0], where[1:]).Find(&us).Error
	if err != nil {
		return nil, fmt.Errorf("TypeUser repository: failed to find all users: %w", err)
	}
	ius := make([]Model, len(us))
	for i, usr := range us {
		usrB := (Model)(&usr)
		ius[i] = usrB
	}
	return ius, nil
}
func (g *RepoTypeUser) Update(u Model) (Model, error) {
	usr := u.(*TypeUser)
	err := g.DB.Save(&usr).Error
	if err != nil {
		return nil, fmt.Errorf("TypeUser repository: failed to update TypeUser: %w", err)
	}
	var iUsr Model = usr
	return iUsr, nil
}
func (g *RepoTypeUser) Delete(id uint) error {
	err := g.DB.Delete(&TypeUser{}, id).Error
	if err != nil {
		return fmt.Errorf("TypeUser repository: failed to delete TypeUser: %w", err)
	}
	return nil
}
func (g *RepoTypeUser) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (g *RepoTypeUser) List(where ...interface{}) (TableHandler, error) {
	var users []TypeUser
	err := g.DB.Where(where[0], where[1:]).Find(&users).Error
	if err != nil {
		return TableHandler{}, fmt.Errorf("TypeUser repository: failed to list users: %w", err)
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, usr := range users {
		tableHandlerMap[i] = map[string]string{
			"id":       usr.GetID(),
			"name":     usr.GetName(),
			"username": usr.GetUsername(),
			"email":    usr.GetEmail(),
		}
	}
	return TableHandler{Rows: tableHandlerMap}, nil
}
func (g *RepoTypeUser) ExecuteCommand(command string, data interface{}) (interface{}, error) {
	commandMap := map[string]func(interface{}) (interface{}, error){
		"create": func(d interface{}) (interface{}, error) {
			return g.Create(d.(Model))
		},
		"findone": func(d interface{}) (interface{}, error) {
			return g.FindOne(d)
		},
		"findall": func(d interface{}) (interface{}, error) {
			return g.FindAll(d)
		},
		"update": func(d interface{}) (interface{}, error) {
			return g.Update(d.(Model))
		},
		"delete": func(d interface{}) (interface{}, error) {
			return nil, g.Delete(d.(uint))
		},
	}
	if fn, ok := commandMap[strings.ToLower(command)]; ok {
		return fn(data)
	}
	return nil, fmt.Errorf("comando desconhecido: %s", command)
}
