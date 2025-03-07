package roles

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
