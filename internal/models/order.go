package models

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type OrderRepo interface {
	Create(o *Order) (*Order, error)
	FindOne(where ...interface{}) (*Order, error)
	FindAll(where ...interface{}) ([]*Order, error)
	Update(o *Order) (*Order, error)
	Delete(id string) error
	Close() error
	List(where ...interface{}) (*TableHandler, error)
}

type OrderRepoImpl struct {
	*gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepoImpl {
	return &OrderRepoImpl{db}
}

func (g *OrderRepoImpl) Create(o *Order) (*Order, error) {
	err := g.DB.Create(o).Error
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (g *OrderRepoImpl) FindOne(where ...interface{}) (*Order, error) {
	var o Order
	err := g.DB.Where(where[0], where[1:]...).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (g *OrderRepoImpl) FindAll(where ...interface{}) ([]*Order, error) {
	var orders []*Order
	err := g.DB.Where(where[0], where[1:]...).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (g *OrderRepoImpl) Update(o *Order) (*Order, error) {
	err := g.DB.Save(o).Error
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (g *OrderRepoImpl) Delete(id string) error {
	err := g.DB.Delete(&Order{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (g *OrderRepoImpl) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (g *OrderRepoImpl) List(where ...interface{}) (*TableHandler, error) {
	var orders []*Order
	err := g.DB.Where(where[0], where[1:]...).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, o := range orders {
		tableHandlerMap[i] = map[string]string{
			"id":           o.ID,
			"order_number": o.OrderNumber,
			"customer_id":  fmt.Sprintf("%d", o.CustomerID),
			"status":       string(o.Status),
			"total_amount": fmt.Sprintf("%.2f", o.TotalAmount),
		}
	}
	return &TableHandler{rows: tableHandlerMap}, nil
}

type Order struct {
	ID                string      `gorm:"type:uuid;primaryKey" json:"id"`
	OrderNumber       string      `gorm:"type:varchar(50);unique;not null" json:"order_number"`
	CustomerID        uint        `gorm:"type:uuid;not null" json:"customer_id"`
	Status            OrderStatus `gorm:"type:varchar(50);not null;default:'draft'" json:"status"`
	OrderDate         time.Time   `gorm:"type:timestamp;not null;default:current_timestamp" json:"order_date"`
	EstimatedDelivery time.Time   `gorm:"type:timestamp" json:"estimated_delivery"`
	ActualDelivery    time.Time   `gorm:"type:timestamp" json:"actual_delivery"`
	TotalAmount       float64     `gorm:"type:decimal(15,2);not null;default:0" json:"total_amount"`
	CreatedAt         time.Time   `gorm:"type:timestamp;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt         time.Time   `gorm:"type:timestamp;not null;default:current_timestamp" json:"updated_at"`
}

type OrderStatus string

const (
	OrderStatusDraft      OrderStatus = "draft"
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

func (o *Order) TableName() string {
	return "orders"
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

func OrderFactory() Order {
	return Order{}
}
