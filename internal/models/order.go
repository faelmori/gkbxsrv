package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepo interface {
	Create(o *Order) (*Order, error)
	FindOne(where ...interface{}) (*Order, error)
	FindAll(where ...interface{}) ([]*Order, error)
	Update(o *Order) (*Order, error)
	Delete(id string) error
	Close() error

	FindAllByCustomerID(customerID int) ([]*Order, error)
	FindAllBySellerID(sellerID int) ([]*Order, error)
}
type OrderRepoImpl struct{ *gorm.DB }

func (o *OrderRepoImpl) Create(order *Order) (*Order, error) {
	err := o.DB.Create(order).Error
	if err != nil {
		return nil, err
	}
	return order, nil
}
func (o *OrderRepoImpl) FindOne(where ...interface{}) (*Order, error) {
	var order Order
	err := o.DB.Where(where[0], where[1:]).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (o *OrderRepoImpl) FindAll(where ...interface{}) ([]*Order, error) {
	var orders []*Order
	err := o.DB.Where(where[0], where[1:]).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}
func (o *OrderRepoImpl) Update(order *Order) (*Order, error) {
	err := o.DB.Save(order).Error
	if err != nil {
		return nil, err
	}
	return order, nil
}
func (o *OrderRepoImpl) Delete(id string) error {
	err := o.DB.Delete(&OrderImpl{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
func (o *OrderRepoImpl) Close() error {
	sqlDB, err := o.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (o *OrderRepoImpl) FindAllByCustomerID(customerID int) ([]*Order, error) {
	var orders []*Order
	err := o.DB.Where("customer_id = ?", customerID).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}
func (o *OrderRepoImpl) FindAllBySellerID(sellerID int) ([]*Order, error) {
	var orders []*Order
	err := o.DB.Where("seller_id = ?", sellerID).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

type Order interface {
	GetID() string
	GetGovernmentID() int
	GetCustomerID() int
	GetSellerID() int
	GetEmployerID() int
	GetOrderDates() OrderDates
	GetOrderValues() OrderValues
	GetOrderShipping() OrderShipping
	GetOrderBilling() OrderBilling
	GetOrderStatus() OrderStatus
	GetOrderItems() []OrderItem
	GetActive() bool
	SetGovernmentID(governmentID int)
	SetCustomerID(customerID int)
	SetSellerID(sellerID int)
	SetEmployerID(employerID int)
	SetOrderDates(orderDates OrderDates)
	SetOrderValues(orderValues OrderValues)
	SetOrderShipping(orderShipping OrderShipping)
	SetOrderBilling(orderBilling OrderBilling)
	SetOrderStatus(orderStatus OrderStatus)
	SetOrderItems(orderItems []OrderItem)
	SetActive(active bool)

	TableName() string
	BeforeCreate(tx *gorm.DB) (err error)
	AfterFind(tx *gorm.DB) (err error)
	AfterSave(tx *gorm.DB) (err error)
	AfterCreate(tx *gorm.DB) (err error)
	AfterUpdate(tx *gorm.DB) (err error)
	AfterDelete(tx *gorm.DB) (err error)
	Sanitize()
	String() string
}

type OrderImpl struct {
	ID            string        `gorm:"primaryKey" json:"id"`
	GovernmentID  int           `json:"government_id"`
	CustomerID    int           `json:"customer_id"`
	SellerID      int           `json:"seller_id"`
	EmployerID    int           `json:"employer_id"`
	OrderDates    OrderDates    `gorm:"embedded" json:"dates"`
	OrderValues   OrderValues   `gorm:"embedded" json:"values"`
	OrderShipping OrderShipping `gorm:"embedded" json:"shipping"`
	OrderBilling  OrderBilling  `gorm:"embedded" json:"billing"`
	OrderStatus   OrderStatus   `gorm:"embedded" json:"status"`
	OrderItems    []OrderItem   `gorm:"foreignKey:OrderID" json:"items"`
	Active        bool          `json:"active"`
}
type OrderDates struct {
	CreatedAt    time.Time `json:"created_at"`
	OrderDate    time.Time `json:"order_date"`
	SellDate     time.Time `json:"sell_date"`
	DueDate      time.Time `json:"due_date"`
	DeliveryDate time.Time `json:"delivery_date"`
}
type OrderValues struct {
	Total      float64 `json:"total"`
	Discount   float64 `json:"discount"`
	Subtotal   float64 `json:"subtotal"`
	Tax        float64 `json:"tax"`
	Shipping   float64 `json:"shipping"`
	GrandTotal float64 `json:"grand_total"`
}
type OrderShipping struct {
	ShippingAddress  string `json:"shipping_address"`
	ShippingRegion   string `json:"shipping_region"`
	ShippingPhone    string `json:"shipping_phone"`
	ShippingEmail    string `json:"shipping_email"`
	ShippingTracking string `json:"shipping_tracking"`
	ShippingCompany  string `json:"shipping_company"`
}
type OrderBilling struct {
	BillingAddress  string `json:"billing_address"`
	BillingRegion   string `json:"billing_region"`
	BillingPhone    string `json:"billing_phone"`
	BillingEmail    string `json:"billing_email"`
	BillingTracking string `json:"billing_tracking"`
	BillingCompany  string `json:"billing_company"`
}
type OrderStatus struct {
	OrderStatus      string `json:"order_status"`
	GovernmentStatus string `json:"government_status"`
	InvoiceStatus    string `json:"invoice_status"`
	PaymentStatus    string `json:"payment_status"`
	ShippingStatus   string `json:"shipping_status"`
	BillingStatus    string `json:"billing_status"`
}
type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   string  `json:"order_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Discount  float64 `json:"discount"`
	Subtotal  float64 `json:"subtotal"`
	Tax       float64 `json:"tax"`
	Total     float64 `json:"total"`
	Active    bool    `json:"active"`
}

func (o *OrderImpl) GetID() string                   { return o.ID }
func (o *OrderImpl) GetGovernmentID() int            { return o.GovernmentID }
func (o *OrderImpl) GetCustomerID() int              { return o.CustomerID }
func (o *OrderImpl) GetSellerID() int                { return o.SellerID }
func (o *OrderImpl) GetEmployerID() int              { return o.EmployerID }
func (o *OrderImpl) GetOrderDates() OrderDates       { return o.OrderDates }
func (o *OrderImpl) GetOrderValues() OrderValues     { return o.OrderValues }
func (o *OrderImpl) GetOrderShipping() OrderShipping { return o.OrderShipping }
func (o *OrderImpl) GetOrderBilling() OrderBilling   { return o.OrderBilling }
func (o *OrderImpl) GetOrderStatus() OrderStatus     { return o.OrderStatus }
func (o *OrderImpl) GetOrderItems() []OrderItem      { return o.OrderItems }
func (o *OrderImpl) GetActive() bool                 { return o.Active }

func (o *OrderImpl) SetGovernmentID(governmentID int) { o.GovernmentID = governmentID }
func (o *OrderImpl) SetCustomerID(customerID int)     { o.CustomerID = customerID }
func (o *OrderImpl) SetSellerID(sellerID int)         { o.SellerID = sellerID }
func (o *OrderImpl) SetEmployerID(employerID int)     { o.EmployerID = employerID }
func (o *OrderImpl) SetOrderDates(orderDatesArg OrderDates) {
	o.OrderDates = OrderDates{
		CreatedAt:    orderDatesArg.CreatedAt,
		OrderDate:    orderDatesArg.OrderDate,
		SellDate:     orderDatesArg.SellDate,
		DueDate:      orderDatesArg.DueDate,
		DeliveryDate: orderDatesArg.DeliveryDate,
	}
}
func (o *OrderImpl) SetOrderValues(orderValuesArg OrderValues) {
	o.OrderValues = OrderValues{
		Total:      orderValuesArg.Total,
		Discount:   orderValuesArg.Discount,
		Subtotal:   orderValuesArg.Subtotal,
		Tax:        orderValuesArg.Tax,
		Shipping:   orderValuesArg.Shipping,
		GrandTotal: orderValuesArg.GrandTotal,
	}
}
func (o *OrderImpl) SetOrderShipping(orderShippingArg OrderShipping) {
	o.OrderShipping = OrderShipping{
		ShippingAddress:  orderShippingArg.ShippingAddress,
		ShippingRegion:   orderShippingArg.ShippingRegion,
		ShippingPhone:    orderShippingArg.ShippingPhone,
		ShippingEmail:    orderShippingArg.ShippingEmail,
		ShippingTracking: orderShippingArg.ShippingTracking,
		ShippingCompany:  orderShippingArg.ShippingCompany,
	}
}
func (o *OrderImpl) SetOrderBilling(orderBillingArg OrderBilling) {
	o.OrderBilling = OrderBilling{
		BillingAddress:  orderBillingArg.BillingAddress,
		BillingRegion:   orderBillingArg.BillingRegion,
		BillingPhone:    orderBillingArg.BillingPhone,
		BillingEmail:    orderBillingArg.BillingEmail,
		BillingTracking: orderBillingArg.BillingTracking,
		BillingCompany:  orderBillingArg.BillingCompany,
	}
}
func (o *OrderImpl) SetOrderStatus(orderStatusArg OrderStatus) {
	o.OrderStatus = OrderStatus{
		OrderStatus:      orderStatusArg.OrderStatus,
		GovernmentStatus: orderStatusArg.GovernmentStatus,
		InvoiceStatus:    orderStatusArg.InvoiceStatus,
		PaymentStatus:    orderStatusArg.PaymentStatus,
		ShippingStatus:   orderStatusArg.ShippingStatus,
		BillingStatus:    orderStatusArg.BillingStatus,
	}
}
func (o *OrderImpl) SetOrderItems(orderItems []OrderItem) { o.OrderItems = orderItems }
func (o *OrderImpl) SetActive(active bool)                { o.Active = active }

func (o *OrderImpl) TableName() string { return "orders" }
func (o *OrderImpl) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = uuid.New().String()
	return nil
}
func (o *OrderImpl) AfterFind(tx *gorm.DB) (err error) {
	o.Sanitize()
	return nil
}
func (o *OrderImpl) AfterSave(tx *gorm.DB) (err error) {
	o.Sanitize()
	return nil
}
func (o *OrderImpl) AfterCreate(tx *gorm.DB) (err error) {
	o.Sanitize()
	return nil
}
func (o *OrderImpl) AfterUpdate(tx *gorm.DB) (err error) {
	o.Sanitize()
	return nil
}
func (o *OrderImpl) AfterDelete(tx *gorm.DB) (err error) {
	o.Sanitize()
	return nil
}
func (o *OrderImpl) Sanitize() {
	// Add any sanitization logic if needed
}
func (o *OrderImpl) String() string {
	return fmt.Sprintf("Order<ID: %s, CustomerID: %d, Total: %.2f>", o.ID, o.CustomerID, o.OrderValues.Total)
}

func NewOrderRepo(db *gorm.DB) OrderRepo { return &OrderRepoImpl{db} }
func OrderFactory(governmentID, customerID, sellerID, employerID int, orderDate, sellDate, dueDate, deliveryDate time.Time, shippingAddress, shippingRegion, shippingPhone, shippingEmail, shippingTracking, shippingCompany, billingAddress, billingRegion, billingPhone, billingEmail, billingTracking, billingCompany, status, governmentStatus, invoiceStatus, paymentStatus, shippingStatus, billingStatus string, total, discount, subtotal, tax, shipping, grandTotal float64, active bool) Order {
	o := OrderImpl{
		GovernmentID: governmentID,
		CustomerID:   customerID,
		SellerID:     sellerID,
		EmployerID:   employerID,
		OrderDates: OrderDates{
			CreatedAt:    time.Now(),
			OrderDate:    orderDate,
			SellDate:     sellDate,
			DueDate:      dueDate,
			DeliveryDate: deliveryDate,
		},
		OrderValues: OrderValues{
			Total:      total,
			Discount:   discount,
			Subtotal:   subtotal,
			Tax:        tax,
			Shipping:   shipping,
			GrandTotal: grandTotal,
		},
		OrderShipping: OrderShipping{
			ShippingAddress:  shippingAddress,
			ShippingRegion:   shippingRegion,
			ShippingPhone:    shippingPhone,
			ShippingEmail:    shippingEmail,
			ShippingTracking: shippingTracking,
			ShippingCompany:  shippingCompany,
		},
		OrderBilling: OrderBilling{
			BillingAddress:  billingAddress,
			BillingRegion:   billingRegion,
			BillingPhone:    billingPhone,
			BillingEmail:    billingEmail,
			BillingTracking: billingTracking,
			BillingCompany:  billingCompany,
		},
		OrderStatus: OrderStatus{
			OrderStatus:      status,
			GovernmentStatus: governmentStatus,
			InvoiceStatus:    invoiceStatus,
			PaymentStatus:    paymentStatus,
			ShippingStatus:   shippingStatus,
			BillingStatus:    billingStatus,
		},
		Active: active,
	}
	return &o
}
