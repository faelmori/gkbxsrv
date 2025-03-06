package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strconv"
)

type Customer interface {
	GetID() string
	GetName() string
	GetAddress() string
	GetRegion() string
	GetPhone() string
	GetEmail() string
	GetScore() int
	GetSeller() int
	GetActive() bool
	SetName(name string)
	SetAddress(address string)
	SetRegion(region string)
	SetPhone(phone string)
	SetEmail(email string)
	SetScore(score int)
	SetSeller(seller int)
	SetActive(active bool)
	BeforeCreate(tx *gorm.DB) (err error)
}
type CustomerImpl struct {
	ID      string `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"name;required" json:"name"`
	Address string `gorm:"address;omitempty" json:"address"`
	Region  string `gorm:"region;omitempty" json:"region"`
	Phone   string `gorm:"phone;omitempty" json:"phone"`
	Email   string `gorm:"email;omitempty" json:"email"`
	Score   int    `gorm:"score;omitempty" json:"score"`
	Seller  int    `gorm:"seller;default:0" json:"seller"`
	Active  bool   `gorm:"active;default:true" json:"active"`
}

func (c *CustomerImpl) TableName() string                    { return "customers" }
func (c *CustomerImpl) BeforeCreate(tx *gorm.DB) (err error) { c.ID = uuid.New().String(); return }
func (c *CustomerImpl) GetID() string                        { return c.ID }
func (c *CustomerImpl) GetName() string                      { return c.Name }
func (c *CustomerImpl) GetAddress() string                   { return c.Address }
func (c *CustomerImpl) GetRegion() string                    { return c.Region }
func (c *CustomerImpl) GetPhone() string                     { return c.Phone }
func (c *CustomerImpl) GetEmail() string                     { return c.Email }
func (c *CustomerImpl) GetScore() int                        { return c.Score }
func (c *CustomerImpl) GetSeller() int                       { return c.Seller }
func (c *CustomerImpl) GetActive() bool                      { return c.Active }
func (c *CustomerImpl) SetName(name string)                  { c.Name = name }
func (c *CustomerImpl) SetAddress(address string)            { c.Address = address }
func (c *CustomerImpl) SetRegion(region string)              { c.Region = region }
func (c *CustomerImpl) SetPhone(phone string)                { c.Phone = phone }
func (c *CustomerImpl) SetEmail(email string)                { c.Email = email }
func (c *CustomerImpl) SetScore(score int)                   { c.Score = score }
func (c *CustomerImpl) SetSeller(seller int)                 { c.Seller = seller }
func (c *CustomerImpl) SetActive(active bool)                { c.Active = active }

type CustomerRepo interface {
	Create(p *Customer) (*Customer, error)
	FindOne(where ...interface{}) (*Customer, error)
	FindAll(where ...interface{}) ([]*Customer, error)
	Update(p *Customer) (*Customer, error)
	Delete(id uint) error
	Close() error
	List(where ...interface{}) (*TableHandler, error)
	ExecuteCommand(command string, data interface{}) (interface{}, error)
}
type CustomerRepoImpl struct{ *gorm.DB }

func (g *CustomerRepoImpl) Create(c *Customer) (*Customer, error) {
	err := g.DB.Create(c).Error
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (g *CustomerRepoImpl) FindOne(where ...interface{}) (*Customer, error) {
	var c Customer
	err := g.DB.Where(where[0], where[1:]).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}
func (g *CustomerRepoImpl) FindAll(where ...interface{}) ([]*Customer, error) {
	var customers []*Customer
	err := g.DB.Where(where[0], where[1:]).Find(&customers).Error
	if err != nil {
		return nil, err
	}
	return customers, nil
}
func (g *CustomerRepoImpl) Update(c *Customer) (*Customer, error) {
	err := g.DB.Save(c).Error
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (g *CustomerRepoImpl) Delete(id uint) error {
	err := g.DB.Delete(&CustomerImpl{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
func (g *CustomerRepoImpl) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (g *CustomerRepoImpl) List(where ...interface{}) (*TableHandler, error) {
	var customers []*CustomerImpl
	err := g.DB.Where(where[0], where[1:]).Find(&customers).Error
	if err != nil {
		return nil, err
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, c := range customers {
		tableHandlerMap[i] = map[string]string{
			"id":      c.ID,
			"name":    c.Name,
			"address": c.Address,
			"region":  c.Region,
			"phone":   c.Phone,
			"email":   c.Email,
			"score":   strconv.Itoa(c.Score),
			"seller":  strconv.Itoa(c.Seller),
			"active":  strconv.FormatBool(c.Active),
		}
	}
	return &TableHandler{tableHandlerMap}, nil
}
func (g *CustomerRepoImpl) ExecuteCommand(command string, data interface{}) (interface{}, error) {
	switch command {
	case "findAll":
		return g.FindAll()
	case "findOne":
		return g.FindOne(data)
	case "create":
		return g.Create(data.(*Customer))
	case "update":
		return g.Update(data.(*Customer))
	case "delete":
		return nil, g.Delete(data.(uint))
	default:
		return nil, nil
	}
}

func NewCustomerRepo(db *gorm.DB) CustomerRepo { return &CustomerRepoImpl{db} }

func CustomerFactory(customerData map[string]interface{}) Customer {
	c := CustomerImpl{}
	if customerData != nil {
		c.Name = customerData["name"].(string)
		c.Address = customerData["address"].(string)
		c.Region = customerData["region"].(string)
		c.Phone = customerData["phone"].(string)
		c.Email = customerData["email"].(string)
		c.Score = customerData["score"].(int)
		c.Seller = customerData["seller"].(int)
		c.Active = customerData["active"].(bool)
	}
	return &c
}
