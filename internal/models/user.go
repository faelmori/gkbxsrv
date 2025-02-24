package models

import (
	"fmt"
	"github.com/faelmori/kbx/mods/logz"

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
	var iUser *userImpl
	if usrIpml, usrIpmlOk := u.(*userImpl); usrIpmlOk {
		iUser = usrIpml
	} else if usrInterface, usrInterfaceOk := u.(User); usrInterfaceOk {
		iUser = usrInterface.getUserObj()
	} else {
		iUser = u.(User).getUserObj()
	}
	err := g.DB.Create(&iUser).Error
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to create UserImpl: %w", err)
	}
	return iUser, nil
}
func (g *UserRepoImpl) FindOne(where ...interface{}) (User, error) {
	var u *userImpl
	err := g.DB.Where(where[0], where[1:]).First(&u).Error
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to find UserImpl: %w", err)
	}
	return u, nil
}
func (g *UserRepoImpl) FindAll(where ...interface{}) ([]User, error) {
	var us []userImpl
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
	usr := u.(*userImpl)
	err := g.DB.Save(&usr).Error // Use Save to update all fields
	if err != nil {
		return nil, fmt.Errorf("UserImpl repository: failed to update UserImpl: %w", err)
	}
	var iUsr User = usr
	return iUsr, nil
}
func (g *UserRepoImpl) Delete(id uint) error { // Delete by ID
	err := g.DB.Delete(&userImpl{}, id).Error // Delete by ID
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
	var users []userImpl
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

	getUserObj() *userImpl
}

func UserFactory(userData map[string]any) User {
	if userData == nil {
		userData = make(map[string]any)
	}
	return convertMapToUser(userData)
}
func convertMapToUser(userData map[string]any) User {
	var u userImpl
	var roleId uint
	if userData["role_id"] == nil {
		roleId = uint(2)
	} else {
		rlID, roleIdErr := strconv.ParseUint(userData["role_id"].(string), 10, 32)
		if roleIdErr != nil {
			_ = logz.WarnLog(fmt.Sprintf("UserImpl factory: failed to convert role_id to int: %v", roleIdErr), "GDBase")
			roleId = uint(2)
		} else {
			roleId = uint(rlID)
		}
	}
	if userData != nil {
		u = userImpl{
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
		u = userImpl{}
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
