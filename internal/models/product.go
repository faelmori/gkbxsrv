package models

import (
	"gorm.io/gorm"
	"strconv"
)

type ProductRepo interface {
	Create(p *Product) (*Product, error)
	FindOne(where ...interface{}) (*Product, error)
	FindAll(where ...interface{}) ([]*Product, error)
	Update(p *Product) (*Product, error)
	Delete(id uint) error
	Close() error
	List(where ...interface{}) (*TableHandler, error)

	FindAllByDepart(depart string) ([]*Product, error)
	FindAllByCategory(category string) ([]*Product, error)
}
type ProductRepoImpl struct {
	*gorm.DB
}

func NewGormProductRepo(db *gorm.DB) *ProductRepoImpl { return &ProductRepoImpl{db} }

func (g *ProductRepoImpl) Create(p *Product) (*Product, error) {
	err := g.DB.Create(p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (g *ProductRepoImpl) FindOne(where ...interface{}) (*Product, error) {
	var p Product
	err := g.DB.Where(where[0], where[1:]).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func (g *ProductRepoImpl) FindAll(where ...interface{}) ([]*Product, error) {
	var products []*Product
	err := g.DB.Where(where[0], where[1:]).Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}
func (g *ProductRepoImpl) Update(p *Product) (*Product, error) {
	err := g.DB.Save(p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (g *ProductRepoImpl) Delete(id uint) error {
	err := g.DB.Delete(&productImpl{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
func (g *ProductRepoImpl) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (g *ProductRepoImpl) List(where ...interface{}) (*TableHandler, error) {
	var products []*productImpl
	err := g.DB.Where(where[0], where[1:]).Find(&products).Error
	if err != nil {
		return nil, err
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, p := range products {
		tableHandlerMap[i] = map[string]string{
			"id":       strconv.Itoa(int(p.ID)),
			"name":     p.Name,
			"depart":   p.Depart,
			"category": p.Category,
			"price":    strconv.FormatFloat(p.Price, 'f', -1, 64),
			"cost":     strconv.FormatFloat(p.Cost, 'f', -1, 64),
			"stock":    strconv.Itoa(p.Stock),
			"reserve":  strconv.Itoa(p.Reserve),
			"balance":  strconv.Itoa(p.Balance),
			"synced":   strconv.FormatBool(p.Synced),
			"lastSync": p.LastSync,
		}
	}
	return &TableHandler{rows: tableHandlerMap}, nil
}
func (g *ProductRepoImpl) FindAllByDepart(depart string) ([]*Product, error) {
	var products []*Product
	err := g.DB.Where("depart = ?", depart).Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}
func (g *ProductRepoImpl) FindAllByCategory(category string) ([]*Product, error) {
	var products []*Product
	err := g.DB.Where("category = ?", category).Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

type Product interface {
	GetID() uint
	GetName() string
	GetDepart() string
	GetCategory() string
	GetPrice() float64
	GetCost() float64
	GetStock() int
	GetReserve() int
	GetBalance() int
	GetSynced() bool
	GetLastSync() string
	SetID(iD uint)
	SetName(name string)
	SetDepart(depart string)
	SetCategory(category string)
	SetPrice(price float64)
	SetCost(cost float64)
	SetStock(stock int)
	SetReserve(reserve int)
	SetBalance(balance int)
	SetSynced(synced bool)
	SetLastSync(lastSync string)

	TableName() string
	Validate() error
	Sanitize()
	Update(name, depart, category string, price, cost float64, stock, reserve, balance int)
	Sync()
	DeductStock(qty int)
	AddStock(qty int)
	DeductReserve(qty int)
	AddReserve(qty int)
	IsAvailable(qty int) bool
	IsLowStock() bool
}

func ProductFactory(name, depart, category string, price, cost float64, stock, reserve, balance int) Product {
	return &productImpl{
		Name:     name,
		Depart:   depart,
		Category: category,
		Price:    price,
		Cost:     cost,
		Stock:    stock,
		Reserve:  reserve,
		Balance:  balance,
	}
}
