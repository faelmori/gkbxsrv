package models

import (
	"github.com/faelmori/gkbxsrv/internal/models"
	"gorm.io/gorm"
)

type CustomerRepo struct{ models.CustomerRepo }
type Customer struct{ models.Customer }

func NewCustomerRepo(db *gorm.DB) *CustomerRepo { return &CustomerRepo{models.NewCustomerRepo(db)} }
func NewCustomer(customerData map[string]interface{}) *Customer {
	return &Customer{models.CustomerFactory(customerData)}
}
