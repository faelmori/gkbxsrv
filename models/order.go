package models

import (
	"github.com/faelmori/gkbxsrv/internal/models"
	"gorm.io/gorm"
	"time"
)

type OrderRepo = models.OrderRepo
type Order = models.Order

func NewOrderRepo(db *gorm.DB) OrderRepo { return models.NewOrderRepo(db) }
func OrderFactory(governmentID, customerID, sellerID, employerID int, orderDate, sellDate, dueDate, deliveryDate time.Time, shippingAddress, shippingRegion, shippingPhone, shippingEmail, shippingTracking, shippingCompany, billingAddress, billingRegion, billingPhone, billingEmail, billingTracking, billingCompany, orderStatus, governmentStatus, invoiceStatus, paymentStatus, shippingStatus, billingStatus string, total, discount, subtotal, tax, shipping, grandTotal float64, active bool) Order {
	return models.OrderFactory()
}
