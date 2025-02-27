package models

import (
	"github.com/faelmori/gkbxsrv/internal/models"
	"gorm.io/gorm"
)

type ProductRepo struct{ models.ProductRepo }
type Product struct{ models.Product }

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{models.NewGormProductRepo(db)} }
func ProductFactory(name, depart, category string, price, cost float64, stock, reserve, balance int) *Product {
	return &Product{models.ProductFactory(name, depart, category, price, cost, stock, reserve, balance)}
}
