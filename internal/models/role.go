package models

import (
	"fmt"
	"github.com/faelmori/kbx/mods/logz"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepo interface {
	Create(u *Role) (*Role, error)
	FindOne(where ...interface{}) (*Role, error)
	FindAll(where ...interface{}) ([]*Role, error)
	Update(u *Role) (*Role, error)
	Delete(id uint) error
	Close() error
	List(where ...interface{}) (*TableHandler, error)
}
type RoleRepoImpl struct{ *gorm.DB }

func (g *RoleRepoImpl) Create(u *Role) (*Role, error) {
	err := g.DB.Create(u).Error
	if err != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("RoleImpl repository: failed to create RoleImpl: %v", err), "GDBase")
	}
	return u, nil
}
func (g *RoleRepoImpl) FindOne(where ...interface{}) (*Role, error) {
	var u Role
	err := g.DB.Where(where[0], where[1:]).First(&u).Error // Use a pointer to u
	if err != nil {
		return nil, logz.ErrorLog(fmt.Sprintf("RoleImpl repository: failed to find RoleImpl: %v", err), "GDBase")
	}
	return &u, nil // Return the pointer
}
func (g *RoleRepoImpl) FindAll(where ...interface{}) ([]*Role, error) {
	var roles []*Role
	err := g.DB.Where(where[0], where[1:]).Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("RoleImpl repository: failed to find all roles: %w", err)
	}
	return roles, nil
}
func (g *RoleRepoImpl) Update(u *Role) (*Role, error) { // Update by ID
	err := g.DB.Save(u).Error // Use Save to update all fields
	if err != nil {
		return nil, fmt.Errorf("RoleImpl repository: failed to update RoleImpl: %w", err)
	}
	return u, nil
}
func (g *RoleRepoImpl) Delete(id uint) error { // Delete by ID
	err := g.DB.Delete(&RoleImpl{}, id).Error // Delete by ID
	if err != nil {
		return fmt.Errorf("RoleImpl repository: failed to delete RoleImpl: %w", err)
	}
	return nil
}
func (g *RoleRepoImpl) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (g *RoleRepoImpl) List(where ...interface{}) (*TableHandler, error) {
	var roles []Role
	err := g.DB.Where(where[0], where[1:]).Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("RoleImpl repository: failed to list roles: %w", err)
	}
	rows := make(map[int]map[string]string)
	for i, rl := range roles {
		if r, ok := rl.(Role); ok {
			rows[i] = map[string]string{
				"id":          r.GetID(),
				"name":        r.GetName(),
				"description": r.GetDescription(),
				"active":      fmt.Sprintf("%t", r.GetActive()),
			}
		}
	}
	return &TableHandler{rows: rows}, nil
}

type Role interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetActive() bool
	SetID(iD string)
	SetName(name string)
	SetDescription(description string)
	SetActive(active bool)
	TableName() string
	BeforeCreate(tx *gorm.DB) (err error)
	AfterFind(tx *gorm.DB) (err error)
	AfterSave(tx *gorm.DB) (err error)
	AfterCreate(tx *gorm.DB) (err error)
	AfterUpdate(tx *gorm.DB) (err error)
	AfterDelete(tx *gorm.DB) (err error)
	Sanitize()
	String() string
	Validate() error
}
type RoleImpl struct {
	ID          string `gorm:"primaryKey" json:"id;omitempty" form:"id;omitempty"`
	Name        string `gorm:"unique" json:"name;omitempty" form:"name;omitempty"`
	Description string `json:"description;omitempty" form:"description;omitempty"`
	Active      bool   `gorm:"default:true" json:"active;omitempty" form:"active;omitempty"`
}

func (u *RoleImpl) GetID() string                     { return u.ID }
func (u *RoleImpl) GetName() string                   { return u.Name }
func (u *RoleImpl) GetDescription() string            { return u.Description }
func (u *RoleImpl) GetActive() bool                   { return u.Active }
func (u *RoleImpl) SetID(iD string)                   { u.ID = iD }
func (u *RoleImpl) SetName(name string)               { u.Name = name }
func (u *RoleImpl) SetDescription(description string) { u.Description = description }
func (u *RoleImpl) SetActive(active bool)             { u.Active = active }

func (u *RoleImpl) TableName() string { return "roles" }
func (u *RoleImpl) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
func (u *RoleImpl) BeforeUpdate(tx *gorm.DB) (err error) {
	return nil
}
func (u *RoleImpl) AfterFind(tx *gorm.DB) (err error) { return nil }
func (u *RoleImpl) AfterSave(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *RoleImpl) AfterCreate(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *RoleImpl) AfterUpdate(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *RoleImpl) AfterDelete(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *RoleImpl) String() string {
	return fmt.Sprintf("Role<%s>", u.ID)
}
func (u *RoleImpl) Sanitize() {
	u.ID = uuid.New().String()
}
func (u *RoleImpl) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("RoleImpl: name is required")
	}
	return nil
}

func RoleFactory() Role {
	var rl = &RoleImpl{
		ID:     uuid.New().String(),
		Active: true,
	}
	return rl
}
func NewRoleRepo(db *gorm.DB) RoleRepo { return &RoleRepoImpl{db} }
