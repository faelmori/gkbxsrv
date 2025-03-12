package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type InventoryRepo interface {
	Create(i *Inventory) (*Inventory, error)
	FindOne(where ...interface{}) (*Inventory, error)
	FindAll(where ...interface{}) ([]*Inventory, error)
	Update(i *Inventory) (*Inventory, error)
	Delete(id string) error
	Close() error
	List(where ...interface{}) (*TableHandler, error)
}

type InventoryRepoImpl struct {
	*gorm.DB
}

func NewInventoryRepo(db *gorm.DB) *InventoryRepoImpl {
	return &InventoryRepoImpl{db}
}

func (g *InventoryRepoImpl) Create(i *Inventory) (*Inventory, error) {
	err := g.DB.Create(i).Error
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (g *InventoryRepoImpl) FindOne(where ...interface{}) (*Inventory, error) {
	var i Inventory
	err := g.DB.Where(where[0], where[1:]...).First(&i).Error
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (g *InventoryRepoImpl) FindAll(where ...interface{}) ([]*Inventory, error) {
	var inventories []*Inventory
	err := g.DB.Where(where[0], where[1:]...).Find(&inventories).Error
	if err != nil {
		return nil, err
	}
	return inventories, nil
}

func (g *InventoryRepoImpl) Update(i *Inventory) (*Inventory, error) {
	err := g.DB.Save(i).Error
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (g *InventoryRepoImpl) Delete(id string) error {
	err := g.DB.Delete(&Inventory{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (g *InventoryRepoImpl) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (g *InventoryRepoImpl) List(where ...interface{}) (*TableHandler, error) {
	var inventories []*Inventory
	err := g.DB.Where(where[0], where[1:]...).Find(&inventories).Error
	if err != nil {
		return nil, err
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, inv := range inventories {
		tableHandlerMap[i] = map[string]string{
			"id":           inv.ID,
			"product_id":   inv.ProductID,
			"warehouse_id": inv.WarehouseID,
			"quantity":     strconv.FormatFloat(inv.Quantity, 'f', 3, 64),
			"status":       string(inv.Status),
		}
	}
	return &TableHandler{rows: tableHandlerMap}, nil
}

type Inventory struct {
	ID            string          `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID     string          `gorm:"type:uuid;not null" json:"product_id"`
	WarehouseID   string          `gorm:"type:uuid;not null" json:"warehouse_id"`
	Quantity      float64         `gorm:"type:decimal(15,3);not null;default:0" json:"quantity"`
	Status        InventoryStatus `gorm:"type:varchar(50);not null;default:'available'" json:"status"`
	LastCountDate time.Time       `gorm:"type:timestamp;not null;default:current_timestamp" json:"last_count_date"`
	CreatedAt     time.Time       `gorm:"type:timestamp;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"type:timestamp;not null;default:current_timestamp" json:"updated_at"`
}

type InventoryStatus string

const (
	InventoryStatusAvailable InventoryStatus = "available"
	InventoryStatusReserved  InventoryStatus = "reserved"
	InventoryStatusDamaged   InventoryStatus = "damaged"
	InventoryStatusExpired   InventoryStatus = "expired"
)

func (i *Inventory) TableName() string {
	return "inventory"
}

func (i *Inventory) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}
