package models

import (
	"fmt"
	//"github.com/faelmori/logz"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ModelList = []interface{}{
	&userImpl{},
	&productImpl{},
	&CustomerImpl{},
	&OrderImpl{},
}

type userImpl struct {
	ID       string `gorm:"required;primaryKey" json:"id" form:"id"`
	Name     string `gorm:"required;not null" json:"name" form:"name"`
	Username string `gorm:"required;unique" json:"username" form:"username"`
	Password string `gorm:"required;not null" json:"password" form:"password"`
	Email    string `gorm:"required;unique" json:"email" form:"email"`
	Phone    string `gorm:"omitempty:default:null" json:"phone" form:"phone"`
	RoleID   uint   `gorm:"omitempty;default:2" json:"role_id" form:"role_id:2"`
	Document string `gorm:"omitempty" json:"document" form:"document"`
	Address  string `gorm:"omitempty" json:"address" form:"address"`
	City     string `gorm:"omitempty" json:"city" form:"city"`
	State    string `gorm:"omitempty" json:"state" form:"state"`
	Country  string `gorm:"omitempty" json:"country" form:"country"`
	Zip      string `gorm:"omitempty" json:"zip" form:"zip"`
	Birth    string `gorm:"omitempty" json:"birth" form:"birth"`
	Avatar   string `gorm:"omitempty" json:"avatar" form:"avatar"`
	Picture  string `gorm:"omitempty" json:"picture" form:"picture"`
	Premium  bool   `gorm:"default:false" json:"premium" form:"premium;"`
	Active   bool   `gorm:"default:true" json:"active" form:"active"`
}

func (u *userImpl) TableName() string { return "users" }
func (u *userImpl) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	if u.Password == "" {
		return ErrPasswordRequired
	}
	hash, hashErr := u.SetPassword(u.Password)
	if hashErr != nil {
		return hashErr
	}
	tx.Statement.Set("password", hash)
	return nil
}
func (u *userImpl) BeforeUpdate(tx *gorm.DB) (err error) {
	var cost int
	var costErr error
	if u.Password == "" {
		return bcrypt.ErrMismatchedHashAndPassword
	}
	if cPass, blPass := tx.Statement.Get("password"); blPass {
		cost, costErr = bcrypt.Cost([]byte(cPass.(string)))
		if costErr != nil || cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
			return bcrypt.InvalidCostError(cost)
		}
	}
	return nil
}
func (u *userImpl) AfterFind(_ *gorm.DB) (err error) { return nil }
func (u *userImpl) AfterSave(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *userImpl) AfterCreate(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *userImpl) AfterUpdate(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *userImpl) AfterDelete(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *userImpl) String() string {
	return fmt.Sprintf("User<ID: %s, Name: %s, Username: %s, Email: %s>", u.ID, u.Name, u.Username, u.Email)
}

func (u *userImpl) SetID(id uuid.UUID)          { u.ID = id.String() }
func (u *userImpl) SetName(name string)         { u.Name = name }
func (u *userImpl) SetUsername(username string) { u.Username = username }
func (u *userImpl) SetEmail(email string)       { u.Email = email }
func (u *userImpl) SetRoleID(roleID uint)       { u.RoleID = roleID }
func (u *userImpl) SetPhone(phone string)       { u.Phone = phone }
func (u *userImpl) SetDocument(document string) { u.Document = document }
func (u *userImpl) SetAddress(address string)   { u.Address = address }
func (u *userImpl) SetCity(city string)         { u.City = city }
func (u *userImpl) SetState(state string)       { u.State = state }
func (u *userImpl) SetCountry(country string)   { u.Country = country }
func (u *userImpl) SetZip(zip string)           { u.Zip = zip }
func (u *userImpl) SetBirth(birth string)       { u.Birth = birth }
func (u *userImpl) SetAvatar(avatar string)     { u.Avatar = avatar }
func (u *userImpl) SetPicture(picture string)   { u.Picture = picture }
func (u *userImpl) SetPremium(premium bool)     { u.Premium = premium }
func (u *userImpl) SetActive(active bool)       { u.Active = active }

func (u *userImpl) GetID() string       { return u.ID }
func (u *userImpl) GetName() string     { return u.Name }
func (u *userImpl) GetUsername() string { return u.Username }
func (u *userImpl) GetEmail() string    { return u.Email }
func (u *userImpl) GetRoleID() uint     { return u.RoleID }
func (u *userImpl) GetPhone() string    { return u.Phone }
func (u *userImpl) GetDocument() string { return u.Document }
func (u *userImpl) GetAddress() string  { return u.Address }
func (u *userImpl) GetCity() string     { return u.City }
func (u *userImpl) GetState() string    { return u.State }
func (u *userImpl) GetCountry() string  { return u.Country }
func (u *userImpl) GetZip() string      { return u.Zip }
func (u *userImpl) GetBirth() string    { return u.Birth }
func (u *userImpl) GetAvatar() string   { return u.Avatar }
func (u *userImpl) GetPicture() string  { return u.Picture }
func (u *userImpl) GetPremium() bool    { return u.Premium }
func (u *userImpl) GetActive() bool     { return u.Active }
func (u *userImpl) SetPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	u.Password = string(bytes)
	return u.Password, err
}
func (u *userImpl) CheckPasswordHash(password string) bool {
	if password == "" {
		//_ = logz.WarnLog("userImpl: password is empty", "GDBase")
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		//_ = logz.DebugLog(fmt.Sprintf("Password check error: %s", err), "GDBase")
	}
	return err == nil
}
func (u *userImpl) Sanitize() {
	u.Password = ""
}
func (u *userImpl) Validate() error {
	if u.Name == "" {
		return &ValidationError{Field: "name", Message: "Name is required"}
	}
	if u.Username == "" {
		return ErrUsernameRequired
	}
	if u.Password == "" {
		return ErrPasswordRequired
	}
	if u.Email == "" {
		return ErrEmailRequired
	}
	return nil
}
func (u *userImpl) getUserObj() *userImpl { return u }

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
