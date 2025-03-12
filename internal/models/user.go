package models

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepo interface {
	Create(u User) (User, error)
	FindOne(where ...interface{}) (User, error)
	FindAll(where ...interface{}) ([]User, error)
	Update(u User) (User, error)
	Delete(id string) error
	Close() error
	List(where ...interface{}) (TableHandler, error)
}

type UserRepoImpl struct{ *gorm.DB }

func NewUserRepo(db *gorm.DB) UserRepo {
	return &UserRepoImpl{db}
}

func (g *UserRepoImpl) Create(u User) (User, error) {
	iUser := u.getUserObj()

	err := g.DB.Create(&iUser).Error
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to create UserImpl: %w", err)
	}
	return iUser, nil
}

func (g *UserRepoImpl) FindOne(where ...interface{}) (User, error) {
	var u UserImpl
	err := g.DB.Where(where[0], where[1:]...).First(&u).Error
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to find UserImpl: %w", err)
	}
	return &u, nil
}

func (g *UserRepoImpl) FindAll(where ...interface{}) ([]User, error) {
	var us []UserImpl
	err := g.DB.Where(where[0], where[1:]...).Find(&us).Error
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to find all users: %w", err)
	}
	ius := make([]User, len(us))
	for i, usr := range us {
		ius[i] = &usr
	}
	return ius, nil
}

func (g *UserRepoImpl) Update(u User) (User, error) {
	usr := u.getUserObj()
	err := g.DB.Save(&usr).Error
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to update UserImpl: %w", err)
	}
	return usr, nil
}

func (g *UserRepoImpl) Delete(id string) error {
	err := g.DB.Delete(&UserImpl{}, id).Error
	if err != nil {
		return fmt.Errorf("UserImpl repository: failed to delete UserImpl: %w", err)
	}
	return nil
}

func (g *UserRepoImpl) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (g *UserRepoImpl) List(where ...interface{}) (TableHandler, error) {
	var users []UserImpl
	err := g.DB.Where(where[0], where[1:]...).Find(&users).Error
	if err != nil {
		return TableHandler{}, fmt.Errorf("UserImpl repository: failed to list users: %w", err)
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, usr := range users {
		tableHandlerMap[i] = map[string]string{
			"id":       usr.ID,
			"name":     usr.Name,
			"username": usr.Username,
			"email":    usr.Email,
		}
	}
	return TableHandler{rows: tableHandlerMap}, nil
}

type User interface {
	GetID() string
	GetName() string
	GetUsername() string
	GetEmail() string
	GetRoleID() uint
	GetPhone() string
	GetDocument() string
	GetAddress() string
	GetCity() string
	GetState() string
	GetCountry() string
	GetZip() string
	GetBirth() string
	GetAvatar() string
	GetPicture() string
	GetActive() bool
	SetName(name string)
	SetUsername(username string)
	SetPassword(password string) error
	SetEmail(email string)
	SetRoleID(roleID uint)
	SetPhone(phone string)
	SetDocument(document string)
	SetAddress(address string)
	SetCity(city string)
	SetState(state string)
	SetCountry(country string)
	SetZip(zip string)
	SetBirth(birth string)
	SetAvatar(avatar string)
	SetPicture(picture string)
	SetActive(active bool)
	CheckPasswordHash(password string) bool
	Sanitize()
	Validate() error
	getUserObj() *UserImpl
}

type UserImpl struct {
	ID       string `gorm:"type:uuid;primaryKey" json:"id"`
	Name     string `gorm:"type:varchar(255);not null" json:"name"`
	Username string `gorm:"type:varchar(255);unique;not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null" json:"password"`
	Email    string `gorm:"type:varchar(255);unique;not null" json:"email"`
	Phone    string `gorm:"type:varchar(20)" json:"phone"`
	RoleID   uint   `gorm:"type:integer;default:2" json:"role_id"`
	Document string `gorm:"type:varchar(20)" json:"document"`
	Address  string `gorm:"type:text" json:"address"`
	City     string `gorm:"type:varchar(100)" json:"city"`
	State    string `gorm:"type:varchar(50)" json:"state"`
	Country  string `gorm:"type:varchar(50)" json:"country"`
	Zip      string `gorm:"type:varchar(20)" json:"zip"`
	Birth    string `gorm:"type:date" json:"birth"`
	Avatar   string `gorm:"type:varchar(255)" json:"avatar"`
	Picture  string `gorm:"type:varchar(255)" json:"picture"`
	Active   bool   `gorm:"type:boolean;default:true" json:"active"`
}

func (u *UserImpl) TableName() string {
	return "users"
}

func (u *UserImpl) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	if u.Password == "" {
		return fmt.Errorf("password is required")
	}
	hash, hashErr := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		return hashErr
	}
	u.Password = string(hash)
	return nil
}

func (u *UserImpl) BeforeUpdate(tx *gorm.DB) (err error) {
	if u.Password == "" {
		return fmt.Errorf("password is required")
	}
	hash, hashErr := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		return hashErr
	}
	u.Password = string(hash)
	return nil
}

func (u *UserImpl) AfterFind(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}

func (u *UserImpl) AfterSave(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}

func (u *UserImpl) AfterCreate(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}

func (u *UserImpl) AfterUpdate(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}

func (u *UserImpl) AfterDelete(tx *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}

func (u *UserImpl) String() string {
	return fmt.Sprintf("User<ID: %s, Name: %s, Username: %s, Email: %s>", u.ID, u.Name, u.Username, u.Email)
}

func (u *UserImpl) SetName(name string) {
	u.Name = name
}

func (u *UserImpl) SetUsername(username string) {
	u.Username = username
}

func (u *UserImpl) SetPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

func (u *UserImpl) SetEmail(email string) {
	u.Email = email
}

func (u *UserImpl) SetRoleID(roleID uint) {
	u.RoleID = roleID
}

func (u *UserImpl) SetPhone(phone string) {
	u.Phone = phone
}

func (u *UserImpl) SetDocument(document string) {
	u.Document = document
}

func (u *UserImpl) SetAddress(address string) {
	u.Address = address
}

func (u *UserImpl) SetCity(city string) {
	u.City = city
}

func (u *UserImpl) SetState(state string) {
	u.State = state
}

func (u *UserImpl) SetCountry(country string) {
	u.Country = country
}

func (u *UserImpl) SetZip(zip string) {
	u.Zip = zip
}

func (u *UserImpl) SetBirth(birth string) {
	u.Birth = birth
}

func (u *UserImpl) SetAvatar(avatar string) {
	u.Avatar = avatar
}

func (u *UserImpl) SetPicture(picture string) {
	u.Picture = picture
}

func (u *UserImpl) SetActive(active bool) {
	u.Active = active
}

func (u *UserImpl) GetID() string {
	return u.ID
}

func (u *UserImpl) GetName() string {
	return u.Name
}

func (u *UserImpl) GetUsername() string {
	return u.Username
}

func (u *UserImpl) GetEmail() string {
	return u.Email
}

func (u *UserImpl) GetRoleID() uint {
	return u.RoleID
}

func (u *UserImpl) GetPhone() string {
	return u.Phone
}

func (u *UserImpl) GetDocument() string {
	return u.Document
}

func (u *UserImpl) GetAddress() string {
	return u.Address
}

func (u *UserImpl) GetCity() string {
	return u.City
}

func (u *UserImpl) GetState() string {
	return u.State
}

func (u *UserImpl) GetCountry() string {
	return u.Country
}

func (u *UserImpl) GetZip() string {
	return u.Zip
}

func (u *UserImpl) GetBirth() string {
	return u.Birth
}

func (u *UserImpl) GetAvatar() string {
	return u.Avatar
}

func (u *UserImpl) GetPicture() string {
	return u.Picture
}

func (u *UserImpl) GetActive() bool {
	return u.Active
}

func (u *UserImpl) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *UserImpl) Sanitize() {
	u.Password = ""
}

func (u *UserImpl) Validate() error {
	if u.Name == "" {
		return &ValidationError{Field: "name", Message: "Name is required"}
	}
	if u.Username == "" {
		return &ValidationError{Field: "username", Message: "Username is required"}
	}
	if u.Password == "" {
		return &ValidationError{Field: "password", Message: "Password is required"}
	}
	if u.Email == "" {
		return &ValidationError{Field: "email", Message: "Email is required"}
	}
	return nil
}

func (u *UserImpl) getUserObj() *UserImpl {
	return u
}

func UserFactory(userData map[string]interface{}) User {
	var u UserImpl
	if userData != nil {
		if name, ok := userData["name"].(string); ok {
			u.Name = name
		}
		if username, ok := userData["username"].(string); ok {
			u.Username = username
		}
		if password, ok := userData["password"].(string); ok {
			u.Password = password
		}
		if email, ok := userData["email"].(string); ok {
			u.Email = email
		}
		if roleID, ok := userData["role_id"].(uint); ok {
			u.RoleID = roleID
		}
		if phone, ok := userData["phone"].(string); ok {
			u.Phone = phone
		}
		if document, ok := userData["document"].(string); ok {
			u.Document = document
		}
		if address, ok := userData["address"].(string); ok {
			u.Address = address
		}
		if city, ok := userData["city"].(string); ok {
			u.City = city
		}
		if state, ok := userData["state"].(string); ok {
			u.State = state
		}
		if country, ok := userData["country"].(string); ok {
			u.Country = country
		}
		if zip, ok := userData["zip"].(string); ok {
			u.Zip = zip
		}
		if birth, ok := userData["birth"].(string); ok {
			u.Birth = birth
		}
		if avatar, ok := userData["avatar"].(string); ok {
			u.Avatar = avatar
		}
		if picture, ok := userData["picture"].(string); ok {
			u.Picture = picture
		}
		if active, ok := userData["active"].(bool); ok {
			u.Active = active
		}
	}
	return &u
}
