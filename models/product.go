package models

import (
	"github.com/faelmori/gkbxsrv/internal/models"
	"gorm.io/gorm"
)

type ProductRepo = models.ProductRepo
type Product = models.Product

func NewProductRepo(db *gorm.DB) ProductRepo { return models.NewGormProductRepo(db) }
func ProductFactory(name, depart, category string, price, cost float64, stock, reserve, balance int) *Product {
	return &Product{}
}
