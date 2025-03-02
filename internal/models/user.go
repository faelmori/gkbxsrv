package models

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	//"github.com/faelmori/logz"

	"gorm.io/gorm"
	"strconv"
)

type UserRepo interface {
	Create(u User) (User, error)
	FindOne(where ...interface{}) (User, error)
	FindAll(where ...interface{}) ([]User, error)
	Update(u User) (User, error)
	Delete(id uint) error
	Close() error
	List(where ...interface{}) (TableHandler, error)
}
type UserRepoImpl struct{ *gorm.DB }

func NewUserRepo(db *gorm.DB) UserRepo {
	usr := UserRepoImpl{db}
	return &usr
}

func (g *UserRepoImpl) Create(u User) (User, error) {
	//if usrIpml, usrIpmlOk := u.(*UserImpl); usrIpmlOk {
	//	iUser = usrIpml
	//} else if usrInterface, usrInterfaceOk := u.(User); usrInterfaceOk {
	//	iUser = usrInterface.getUserObj()
	//} else {
	//	iUser = u.(User).getUserObj()
	//}

	iUser := u.getUserObj()

	err := g.DB.Create(&iUser).Error

	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to create UserImpl: %w", err)
	}
	return iUser, nil
}
func (g *UserRepoImpl) FindOne(where ...interface{}) (User, error) {
	var u *UserImpl
	err := g.DB.Where(where[0], where[1:]).First(&u).Error
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to find UserImpl: %w", err)
	}
	return u, nil
}
func (g *UserRepoImpl) FindAll(where ...interface{}) ([]User, error) {
	var us []UserImpl
	err := g.DB.Where(where[0], where[1:]).Find(&us).Error
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to find all users: %w", err)
	}
	ius := make([]User, len(us))
	for i, usr := range us {
		usrB := (User)(&usr)
		ius[i] = usrB
	}
	return ius, nil
}
func (g *UserRepoImpl) Update(u User) (User, error) { // Update by ID
	usr := u.(*UserImpl)
	err := g.DB.Save(&usr).Error // Use Save to update all fields
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to update UserImpl: %w", err)
	}
	var iUsr User = usr
	return iUsr, nil
}
func (g *UserRepoImpl) Delete(id uint) error { // Delete by ID
	err := g.DB.Delete(&UserImpl{}, id).Error // Delete by ID
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
	err := g.DB.Where(where[0], where[1:]).Find(&users).Error
	if err != nil {
		return TableHandler{}, fmt.Errorf("UserImpl repository: failed to list users: %w", err)
	}
	tableHandlerMap := make(map[int]map[string]string)
	for i, usr := range users {
		tableHandlerMap[i] = map[string]string{
			"id":       usr.GetID(),
			"name":     usr.GetName(),
			"username": usr.GetUsername(),
			"email":    usr.GetEmail(),
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
	GetPremium() bool
	GetActive() bool
	SetName(Name string)
	SetUsername(Username string)
	SetPassword(Password string) (string, error)

	SetEmail(Email string)
	SetRoleID(RoleID uint)
	SetPhone(Phone string)
	SetDocument(Document string)
	SetAddress(Address string)
	SetCity(City string)
	SetState(State string)
	SetCountry(Country string)
	SetZip(Zip string)
	SetBirth(Birth string)
	SetAvatar(Avatar string)
	SetPicture(Picture string)
	SetPremium(Premium bool)
	SetActive(Active bool)
	CheckPasswordHash(password string) bool
	Sanitize()
	Validate() error

	getUserObj() *UserImpl
}

type UserImpl struct {
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

func (u *UserImpl) TableName() string { return "users" }
func (u *UserImpl) BeforeCreate(tx *gorm.DB) (err error) {
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
func (u *UserImpl) BeforeUpdate(tx *gorm.DB) (err error) {
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
func (u *UserImpl) AfterFind(_ *gorm.DB) (err error) { return nil }
func (u *UserImpl) AfterSave(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *UserImpl) AfterCreate(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *UserImpl) AfterUpdate(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *UserImpl) AfterDelete(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *UserImpl) String() string {
	return fmt.Sprintf("User<ID: %s, Name: %s, Username: %s, Email: %s>", u.ID, u.Name, u.Username, u.Email)
}
func (u *UserImpl) SetID(id uuid.UUID)          { u.ID = id.String() }
func (u *UserImpl) SetName(name string)         { u.Name = name }
func (u *UserImpl) SetUsername(username string) { u.Username = username }
func (u *UserImpl) SetEmail(email string)       { u.Email = email }
func (u *UserImpl) SetRoleID(roleID uint)       { u.RoleID = roleID }
func (u *UserImpl) SetPhone(phone string)       { u.Phone = phone }
func (u *UserImpl) SetDocument(document string) { u.Document = document }
func (u *UserImpl) SetAddress(address string)   { u.Address = address }
func (u *UserImpl) SetCity(city string)         { u.City = city }
func (u *UserImpl) SetState(state string)       { u.State = state }
func (u *UserImpl) SetCountry(country string)   { u.Country = country }
func (u *UserImpl) SetZip(zip string)           { u.Zip = zip }
func (u *UserImpl) SetBirth(birth string)       { u.Birth = birth }
func (u *UserImpl) SetAvatar(avatar string)     { u.Avatar = avatar }
func (u *UserImpl) SetPicture(picture string)   { u.Picture = picture }
func (u *UserImpl) SetPremium(premium bool)     { u.Premium = premium }
func (u *UserImpl) SetActive(active bool)       { u.Active = active }

func (u *UserImpl) GetID() string       { return u.ID }
func (u *UserImpl) GetName() string     { return u.Name }
func (u *UserImpl) GetUsername() string { return u.Username }
func (u *UserImpl) GetEmail() string    { return u.Email }
func (u *UserImpl) GetRoleID() uint     { return u.RoleID }
func (u *UserImpl) GetPhone() string    { return u.Phone }
func (u *UserImpl) GetDocument() string { return u.Document }
func (u *UserImpl) GetAddress() string  { return u.Address }
func (u *UserImpl) GetCity() string     { return u.City }
func (u *UserImpl) GetState() string    { return u.State }
func (u *UserImpl) GetCountry() string  { return u.Country }
func (u *UserImpl) GetZip() string      { return u.Zip }
func (u *UserImpl) GetBirth() string    { return u.Birth }
func (u *UserImpl) GetAvatar() string   { return u.Avatar }
func (u *UserImpl) GetPicture() string  { return u.Picture }
func (u *UserImpl) GetPremium() bool    { return u.Premium }
func (u *UserImpl) GetActive() bool     { return u.Active }
func (u *UserImpl) SetPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	u.Password = string(bytes)
	return u.Password, err
}
func (u *UserImpl) CheckPasswordHash(password string) bool {
	if password == "" {
		//_ = logz.WarnLog("UserImpl: password is empty", "GDBase")
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		//_ = logz.DebugLog(fmt.Sprintf("Password check error: %s", err), "GDBase")
	}
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
func (u *UserImpl) getUserObj() *UserImpl { return u }

func UserFactory(userData map[string]any) User {
	if userData == nil {
		userData = make(map[string]any)
	}
	return convertMapToUser(userData)
}
func convertMapToUser(userData map[string]any) User {
	var u UserImpl
	var roleId uint
	if userData["role_id"] == nil {
		roleId = uint(2)
	} else {
		rlID, roleIdErr := strconv.ParseUint(userData["role_id"].(string), 10, 32)
		if roleIdErr != nil {
			//_ = logz.WarnLog(fmt.Sprintf("UserImpl factory: failed to convert role_id to int: %v", roleIdErr), "GDBase")
			roleId = uint(2)
		} else {
			roleId = uint(rlID)
		}
	}
	if userData != nil {
		u = UserImpl{
			Name:     convStrField(userData["name"]),
			Username: convStrField(userData["username"]),
			Password: convStrField(userData["password"]),
			Email:    convStrField(userData["email"]),
			RoleID:   roleId,
			Phone:    convStrField(userData["phone"]),
			Document: convStrField(userData["document"]),
			Address:  convStrField(userData["address"]),
			City:     convStrField(userData["city"]),
			State:    convStrField(userData["state"]),
			Country:  convStrField(userData["country"]),
			Zip:      convStrField(userData["zip"]),
			Birth:    convStrField(userData["birth"]),
			Avatar:   convStrField(userData["avatar"]),
			Picture:  convStrField(userData["picture"]),
			Premium:  convStrField(userData["premium"]) == "true",
			Active:   convStrField(userData["active"]) == "true",
		}
	} else {
		u = UserImpl{}
	}
	return &u
}
func convStrField(field interface{}) string {
	if field == nil {
		return ""
	}
	var strField string
	var strOk bool
	strField, strOk = field.(string)
	if !strOk {
		if stField, stOk := field.(*string); stOk {
			strField = *stField
		} else {
			if intField, intOk := field.(int); intOk {
				strField = strconv.Itoa(intField)
			} else {
				strField = ""
			}
		}
	}
	return strField
}
