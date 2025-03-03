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

type productImpl struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	Name     string  `json:"name" gorm:"not null"`
	Depart   string  `json:"depart" gorm:"not null"`
	Category string  `json:"category" gorm:"not null"`
	Price    float64 `json:"price" gorm:"not null"`
	Cost     float64 `json:"cost" gorm:"not null"`
	Stock    int     `json:"stock" gorm:"not null"`
	Reserve  int     `json:"reserve" gorm:"not null"`
	Balance  int     `json:"balance" gorm:"not null"`
	Synced   bool    `json:"synced" gorm:"not null"`
	LastSync string  `json:"last_sync" gorm:"not null"`
}

func (p *productImpl) GetID() uint                 { return p.ID }
func (p *productImpl) GetName() string             { return p.Name }
func (p *productImpl) GetDepart() string           { return p.Depart }
func (p *productImpl) GetCategory() string         { return p.Category }
func (p *productImpl) GetPrice() float64           { return p.Price }
func (p *productImpl) GetCost() float64            { return p.Cost }
func (p *productImpl) GetStock() int               { return p.Stock }
func (p *productImpl) GetReserve() int             { return p.Reserve }
func (p *productImpl) GetBalance() int             { return p.Balance }
func (p *productImpl) GetSynced() bool             { return p.Synced }
func (p *productImpl) GetLastSync() string         { return p.LastSync }
func (p *productImpl) SetID(iD uint)               { p.ID = iD }
func (p *productImpl) SetName(name string)         { p.Name = name }
func (p *productImpl) SetDepart(depart string)     { p.Depart = depart }
func (p *productImpl) SetCategory(category string) { p.Category = category }
func (p *productImpl) SetPrice(price float64)      { p.Price = price }
func (p *productImpl) SetCost(cost float64)        { p.Cost = cost }
func (p *productImpl) SetStock(stock int)          { p.Stock = stock }
func (p *productImpl) SetReserve(reserve int)      { p.Reserve = reserve }
func (p *productImpl) SetBalance(balance int)      { p.Balance = balance }
func (p *productImpl) SetSynced(synced bool)       { p.Synced = synced }
func (p *productImpl) SetLastSync(lastSync string) { p.LastSync = lastSync }
func (p *productImpl) TableName() string {
	return "products"
}
func (p *productImpl) Validate() error {
	if p.Name == "" {
		return &ValidationError{Field: "name", Message: "Name is required"}
	}
	if p.Depart == "" {
		return &ValidationError{Field: "depart", Message: "Depart is required"}
	}
	if p.Category == "" {
		return &ValidationError{Field: "category", Message: "Category is required"}
	}
	if p.Price == 0 {
		return &ValidationError{Field: "price", Message: "Price is required"}
	}
	if p.Cost == 0 {
		return &ValidationError{Field: "cost", Message: "Cost is required"}
	}
	if p.Stock == 0 {
		return &ValidationError{Field: "stock", Message: "Stock is required"}
	}
	if p.Reserve == 0 {
		return &ValidationError{Field: "reserve", Message: "Reserve is required"}
	}
	if p.Balance == 0 {
		return &ValidationError{Field: "balance", Message: "Balance is required"}
	}

	return nil
}
func (p *productImpl) Sanitize() {
	p.Synced = false
	p.LastSync = ""
}
func (p *productImpl) Update(name, depart, category string, price, cost float64, stock, reserve, balance int) {
	p.Name = name
	p.Depart = depart
	p.Category = category
	p.Price = price
	p.Cost = cost
	p.Stock = stock
	p.Reserve = reserve
	p.Balance = balance
}
func (p *productImpl) Sync() {
	p.Synced = true
	p.LastSync = "2021-06-01"
}
func (p *productImpl) DeductStock(qty int) {
	p.Stock -= qty
	p.Balance = p.Stock - p.Reserve
}
func (p *productImpl) AddStock(qty int) {
	p.Stock += qty
	p.Balance = p.Stock - p.Reserve
}
func (p *productImpl) DeductReserve(qty int) {
	p.Reserve -= qty
	p.Balance = p.Stock - p.Reserve
}
func (p *productImpl) AddReserve(qty int) {
	p.Reserve += qty
	p.Balance = p.Stock - p.Reserve
}
func (p *productImpl) IsAvailable(qty int) bool {
	return p.Balance >= qty
}
func (p *productImpl) IsLowStock() bool {
	return p.Balance <= 10
}
