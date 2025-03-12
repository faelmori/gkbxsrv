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

func NewGormProductRepo(db *gorm.DB) *ProductRepoImpl {
	return &ProductRepoImpl{db}
}

func (g *ProductRepoImpl) Create(p *Product) (*Product, error) {
	err := g.DB.Create(p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (g *ProductRepoImpl) FindOne(where ...interface{}) (*Product, error) {
	var p Product
	err := g.DB.Where(where[0], where[1:]...).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (g *ProductRepoImpl) FindAll(where ...interface{}) ([]*Product, error) {
	var products []*Product
	err := g.DB.Where(where[0], where[1:]...).Find(&products).Error
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
	err := g.DB.Delete(&Product{}, id).Error
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
	var products []*Product
	err := g.DB.Where(where[0], where[1:]...).Find(&products).Error
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

type Product struct {
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

func (p *Product) TableName() string {
	return "products"
}

func (p *Product) Validate() error {
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

func (p *Product) Sanitize() {
	p.Synced = false
	p.LastSync = ""
}

func (p *Product) Update(name, depart, category string, price, cost float64, stock, reserve, balance int) {
	p.Name = name
	p.Depart = depart
	p.Category = category
	p.Price = price
	p.Cost = cost
	p.Stock = stock
	p.Reserve = reserve
	p.Balance = balance
}

func (p *Product) Sync() {
	p.Synced = true
	p.LastSync = "2021-06-01"
}

func (p *Product) DeductStock(qty int) {
	p.Stock -= qty
	p.Balance = p.Stock - p.Reserve
}

func (p *Product) AddStock(qty int) {
	p.Stock += qty
	p.Balance = p.Stock - p.Reserve
}

func (p *Product) DeductReserve(qty int) {
	p.Reserve -= qty
	p.Balance = p.Stock - p.Reserve
}

func (p *Product) AddReserve(qty int) {
	p.Reserve += qty
	p.Balance = p.Stock - p.Reserve
}

func (p *Product) IsAvailable(qty int) bool {
	return p.Balance >= qty
}

func (p *Product) IsLowStock() bool {
	return p.Balance <= 10
}
