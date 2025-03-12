package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strconv"
)

type WarehouseRepo interface {
	Create(w *Warehouse) (*Warehouse, error)
	FindOne(where ...interface{}) (*Warehouse, error)
	FindAll(where ...interface{}) ([]*Warehouse, error)
	Update(w *Warehouse) (*Warehouse, error)
	Delete(id string) error
	Close() error
	List(where ...interface{}) (*TableHandler, error)
}

type WarehouseRepoImpl struct {
	*gorm.DB
}

func NewWarehouseRepo(db *gorm.DB) *WarehouseRepoImpl {
	return &WarehouseRepoImpl{db}
}

func (g *WarehouseRepoImpl) Create(w *Warehouse) (*Warehouse, error) {
	err := g.DB.Create(w).Error
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (g *WarehouseRepoImpl) FindOne(where ...interface{}) (*Warehouse, error) {
	var w Warehouse
	err := g.DB.Where(where[0], where[1:]...).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (g *WarehouseRepoImpl) FindAll(where ...interface{}) ([]*Warehouse, error) {
	var warehouses []*Warehouse
	err := g.DB.Where(where[0], where[1:]...).Find(&warehouses).Error
	if err != nil {
		return nil, err
	}
	return warehouses, nil
}

func (g *WarehouseRepoImpl) Update(w *Warehouse) (*Warehouse, error) {
	err := g.DB.Save(w).Error
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (g *WarehouseRepoImpl) Delete(id string) error {
	err := g.DB.Delete(&Warehouse{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (g *WarehouseRepoImpl) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (g *WarehouseRepoImpl) List(where ...interface{}) (*TableHandler, error) {
	var warehouses []*Warehouse
	err := g.DB.Where(where[0], where[1:]...).Find(&warehouses).Error
	if err != nil {
		return nil, err
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, w := range warehouses {
		tableHandlerMap[i] = map[string]string{
			"id":      w.ID,
			"code":    w.Code,
			"name":    w.Name,
			"address": w.Address,
			"city":    w.City,
			"state":   w.State,
			"country": w.Country,
			"active":  strconv.FormatBool(w.Active),
		}
	}
	return &TableHandler{rows: tableHandlerMap}, nil
}

type Warehouse struct {
	ID         string `gorm:"type:uuid;primaryKey" json:"id"`
	Code       string `gorm:"type:varchar(50);unique;not null" json:"code"`
	Name       string `gorm:"type:varchar(255);not null" json:"name"`
	Address    string `gorm:"type:text" json:"address"`
	City       string `gorm:"type:varchar(100)" json:"city"`
	State      string `gorm:"type:varchar(50)" json:"state"`
	Country    string `gorm:"type:varchar(50);default:'Brasil'" json:"country"`
	PostalCode string `gorm:"type:varchar(20)" json:"postal_code"`
	Active     bool   `gorm:"type:boolean;default:true" json:"active"`
}

func (w *Warehouse) TableName() string {
	return "warehouses"
}

func (w *Warehouse) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}
