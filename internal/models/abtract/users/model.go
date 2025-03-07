package users

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
}
