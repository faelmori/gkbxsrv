package roles

import (
	"gorm.io/gorm"
)

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
