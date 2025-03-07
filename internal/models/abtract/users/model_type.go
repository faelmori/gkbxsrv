package users

import (
	"fmt"
	. "github.com/faelmori/gkbxsrv/internal/models/abtract/commons"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strconv"
)

type TypeUser struct {
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

func (u *TypeUser) TableName() string { return "users" }
func (u *TypeUser) BeforeCreate(tx *gorm.DB) (err error) {
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
func (u *TypeUser) BeforeUpdate(tx *gorm.DB) (err error) {
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
func (u *TypeUser) AfterFind(_ *gorm.DB) (err error) { return nil }
func (u *TypeUser) AfterSave(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *TypeUser) AfterCreate(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *TypeUser) AfterUpdate(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *TypeUser) AfterDelete(_ *gorm.DB) (err error) {
	u.Sanitize()
	return nil
}
func (u *TypeUser) String() string {
	return fmt.Sprintf("User<ID: %s, Name: %s, Username: %s, Email: %s>", u.ID, u.Name, u.Username, u.Email)
}
func (u *TypeUser) SetID(id uuid.UUID)          { u.ID = id.String() }
func (u *TypeUser) SetName(name string)         { u.Name = name }
func (u *TypeUser) SetUsername(username string) { u.Username = username }
func (u *TypeUser) SetEmail(email string)       { u.Email = email }
func (u *TypeUser) SetRoleID(roleID uint)       { u.RoleID = roleID }
func (u *TypeUser) SetPhone(phone string)       { u.Phone = phone }
func (u *TypeUser) SetDocument(document string) { u.Document = document }
func (u *TypeUser) SetAddress(address string)   { u.Address = address }
func (u *TypeUser) SetCity(city string)         { u.City = city }
func (u *TypeUser) SetState(state string)       { u.State = state }
func (u *TypeUser) SetCountry(country string)   { u.Country = country }
func (u *TypeUser) SetZip(zip string)           { u.Zip = zip }
func (u *TypeUser) SetBirth(birth string)       { u.Birth = birth }
func (u *TypeUser) SetAvatar(avatar string)     { u.Avatar = avatar }
func (u *TypeUser) SetPicture(picture string)   { u.Picture = picture }
func (u *TypeUser) SetPremium(premium bool)     { u.Premium = premium }
func (u *TypeUser) SetActive(active bool)       { u.Active = active }

func (u *TypeUser) GetID() string       { return u.ID }
func (u *TypeUser) GetName() string     { return u.Name }
func (u *TypeUser) GetUsername() string { return u.Username }
func (u *TypeUser) GetEmail() string    { return u.Email }
func (u *TypeUser) GetRoleID() uint     { return u.RoleID }
func (u *TypeUser) GetPhone() string    { return u.Phone }
func (u *TypeUser) GetDocument() string { return u.Document }
func (u *TypeUser) GetAddress() string  { return u.Address }
func (u *TypeUser) GetCity() string     { return u.City }
func (u *TypeUser) GetState() string    { return u.State }
func (u *TypeUser) GetCountry() string  { return u.Country }
func (u *TypeUser) GetZip() string      { return u.Zip }
func (u *TypeUser) GetBirth() string    { return u.Birth }
func (u *TypeUser) GetAvatar() string   { return u.Avatar }
func (u *TypeUser) GetPicture() string  { return u.Picture }
func (u *TypeUser) GetPremium() bool    { return u.Premium }
func (u *TypeUser) GetActive() bool     { return u.Active }
func (u *TypeUser) SetPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	u.Password = string(bytes)
	return u.Password, err
}
func (u *TypeUser) CheckPasswordHash(password string) bool {
	if password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
func (u *TypeUser) Sanitize() {
	u.Password = ""
}
func (u *TypeUser) Validate() error {
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
func (u *TypeUser) getUserObj() *TypeUser { return u }

func UserFactory(userData map[string]any) Model {
	if userData == nil {
		userData = make(map[string]any)
	}
	return convertMapToUser(userData)
}
func convertMapToUser(userData map[string]any) Model {
	var u TypeUser
	var roleId uint
	if userData["role_id"] == nil {
		roleId = uint(2)
	} else {
		rlID, roleIdErr := strconv.ParseUint(userData["role_id"].(string), 10, 32)
		if roleIdErr != nil {
			roleId = uint(2)
		} else {
			roleId = uint(rlID)
		}
	}
	if userData != nil {
		u = TypeUser{
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
		u = TypeUser{}
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
