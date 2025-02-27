package models

import (
	"github.com/faelmori/gkbxsrv/internal/models"
	"gorm.io/gorm"
	"time"
)

type OrderRepo struct{ models.OrderRepo }
type Order struct{ models.Order }

func NewOrderRepo(db *gorm.DB) *OrderRepo { return &OrderRepo{models.NewOrderRepo(db)} }
func OrderFactory(governmentID, customerID, sellerID, employerID int, orderDate, sellDate, dueDate, deliveryDate time.Time, shippingAddress, shippingRegion, shippingPhone, shippingEmail, shippingTracking, shippingCompany, billingAddress, billingRegion, billingPhone, billingEmail, billingTracking, billingCompany, orderStatus, governmentStatus, invoiceStatus, paymentStatus, shippingStatus, billingStatus string, total, discount, subtotal, tax, shipping, grandTotal float64, active bool) *Order {
	return &Order{models.OrderFactory(governmentID, customerID, sellerID, employerID, orderDate, sellDate, dueDate, deliveryDate, shippingAddress, shippingRegion, shippingPhone, shippingEmail, shippingTracking, shippingCompany, billingAddress, billingRegion, billingPhone, billingEmail, billingTracking, billingCompany, orderStatus, governmentStatus, invoiceStatus, paymentStatus, shippingStatus, billingStatus, total, discount, subtotal, tax, shipping, grandTotal, active)}
}
