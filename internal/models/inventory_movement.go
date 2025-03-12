package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type InventoryMovementRepo interface {
	Create(im *InventoryMovement) (*InventoryMovement, error)
	FindOne(where ...interface{}) (*InventoryMovement, error)
	FindAll(where ...interface{}) ([]*InventoryMovement, error)
	Update(im *InventoryMovement) (*InventoryMovement, error)
	Delete(id string) error
	Close() error
	List(where ...interface{}) (*TableHandler, error)
}

type InventoryMovementRepoImpl struct {
	*gorm.DB
}

func NewInventoryMovementRepo(db *gorm.DB) *InventoryMovementRepoImpl {
	return &InventoryMovementRepoImpl{db}
}

func (g *InventoryMovementRepoImpl) Create(im *InventoryMovement) (*InventoryMovement, error) {
	err := g.DB.Create(im).Error
	if err != nil {
		return nil, err
	}
	return im, nil
}

func (g *InventoryMovementRepoImpl) FindOne(where ...interface{}) (*InventoryMovement, error) {
	var im InventoryMovement
	err := g.DB.Where(where[0], where[1:]...).First(&im).Error
	if err != nil {
		return nil, err
	}
	return &im, nil
}

func (g *InventoryMovementRepoImpl) FindAll(where ...interface{}) ([]*InventoryMovement, error) {
	var inventoryMovements []*InventoryMovement
	err := g.DB.Where(where[0], where[1:]...).Find(&inventoryMovements).Error
	if err != nil {
		return nil, err
	}
	return inventoryMovements, nil
}

func (g *InventoryMovementRepoImpl) Update(im *InventoryMovement) (*InventoryMovement, error) {
	err := g.DB.Save(im).Error
	if err != nil {
		return nil, err
	}
	return im, nil
}

func (g *InventoryMovementRepoImpl) Delete(id string) error {
	err := g.DB.Delete(&InventoryMovement{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (g *InventoryMovementRepoImpl) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (g *InventoryMovementRepoImpl) List(where ...interface{}) (*TableHandler, error) {
	var inventoryMovements []*InventoryMovement
	err := g.DB.Where(where[0], where[1:]...).Find(&inventoryMovements).Error
	if err != nil {
		return nil, err
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, im := range inventoryMovements {
		tableHandlerMap[i] = map[string]string{
			"id":           im.ID,
			"inventory_id": im.InventoryID,
			"product_id":   im.ProductID,
			"quantity":     strconv.FormatFloat(im.Quantity, 'f', 3, 64),
			"type":         im.MovementType,
		}
	}
	return &TableHandler{rows: tableHandlerMap}, nil
}

type InventoryMovement struct {
	ID                string    `gorm:"type:uuid;primaryKey" json:"id"`
	InventoryID       string    `gorm:"type:uuid;not null" json:"inventory_id"`
	ProductID         string    `gorm:"type:uuid;not null" json:"product_id"`
	Quantity          float64   `gorm:"type:decimal(15,3);not null" json:"quantity"`
	MovementType      string    `gorm:"type:varchar(50);not null" json:"movement_type"` // entrada, saída, ajuste, transferência
	ReferenceDocument string    `gorm:"type:varchar(100)" json:"reference_document"`
	Reason            string    `gorm:"type:text" json:"reason"`
	CreatedAt         time.Time `gorm:"type:timestamp;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt         time.Time `gorm:"type:timestamp;not null;default:current_timestamp" json:"updated_at"`
}

func (im *InventoryMovement) TableName() string {
	return "inventory_movements"
}

func (im *InventoryMovement) BeforeCreate(tx *gorm.DB) (err error) {
	if im.ID == "" {
		im.ID = uuid.New().String()
	}
	return nil
}
