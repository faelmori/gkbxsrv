package users

import (
	//"github.com/faelmori/xtui/types"
	"github.com/go-playground/validator/v10"
	"sync"
)

var (
	mu       sync.Mutex
	dskCache = make(map[string]string)
	validate *validator.Validate
)

type UserForm struct{}

/*
	func (u *UserForm) GetFields(userData map[string]string) []types.TuizInput {
		mu.Lock()
		if len(userData) > 0 {
			for k, v := range userData {
				dskCache[k] = v
			}
		} else {
			if len(dskCache) > 0 {
				for k, v := range dskCache {
					userData[k] = v
				}
			} else {
				dskCache = make(map[string]string)
			}
		}
		mu.Unlock()
		return []types.TuizInput{
			createInput("Enter the username", "text", userData["username"], true, 0, 50, "Username must be between 5 and 50 characters", u.validateUsername),
			createInput("Enter the email", "email", userData["email"], true, 0, 50, "Email must be between 5 and 50 characters", u.validateEmail),
			createInput("Enter the password", "password", userData["password"], true, 0, 13, "Password must be between 8 and 13 characters, and contain at least one uppercase letter, one lowercase letter, one number, and one special character", u.validatePassword),
			createInput("Enter the name", "text", userData["name"], true, 0, 50, "Name must be between 5 and 50 characters and contain only letters", u.validateName),
			createInput("Enter the role", "text", userData["role"], false, 0, 10, "Role must be between 1 and 10 characters", u.validateRole),
			createInput("Enter the birth", "date", userData["birth"], false, 0, 10, "Birth must be in the format YYYY-MM-DD", u.validateBirth),
			createInput("Enter the phone", "text", userData["phone"], false, 0, 15, "Phone must be between 10 and 15 characters and contain only numbers", u.validatePhone),
			createInput("Enter the document", "text", userData["document"], false, 0, 15, "Document must be between 10 and 15 characters and contain only numbers", u.validateDocument),
			createInput("Enter the address", "text", userData["address"], false, 0, 150, "Address must be between 5 and 150 characters", u.validateAddress),
			createInput("Enter the city", "text", userData["city"], false, 0, 50, "City must be between 2 and 50 characters", u.validateCity),
			createInput("Enter the state", "text", userData["state"], false, 0, 50, "State must be between 2 and 50 characters", u.validateState),
			createInput("Enter the country", "text", userData["country"], false, 0, 50, "Country must be between 2 and 50 characters", u.validateCountry),
			createInput("Enter the zip code", "text", userData["zip"], false, 0, 10, "Zip code must be between 5 and 10 characters", u.validateZip),
			createInput("Enter the avatar URL", "url", userData["avatar"], false, 0, 200, "Avatar URL must be between 5 and 200 characters", u.validateURL),
			createInput("Enter the picture URL", "url", userData["picture"], false, 0, 200, "Picture URL must be between 5 and 200 characters", u.validateURL),
		}
	}
*/
func (u *UserForm) validateUsername(usernameVal string) error {
	//if err := validate.Var(usernameVal, "required,min=5,max=50"); err != nil {
	//	return fmt.Errorf("username must be between 5 and 50 characters")
	//}
	return nil
}
func (u *UserForm) validateEmail(emailVal string) error {
	//if err := validate.Var(emailVal, "required,email,min=5,max=50"); err != nil {
	//	return fmt.Errorf("email must be between 5 and 50 characters")
	//}
	return nil
}
func (u *UserForm) validatePassword(passwordVal string) error {
	//if err := validate.Var(passwordVal, "required,min=8,max=13,containsany=!@#$%^&*,containsany=ABCDEFGHIJKLMNOPQRSTUVWXYZ,containsany=abcdefghijklmnopqrstuvwxyz,containsany=0123456789"); err != nil {
	//	return fmt.Errorf("password must be between 8 and 13 characters, and contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	//}
	return nil
}
func (u *UserForm) validateName(nameVal string) error {
	//if err := validate.Var(nameVal, "required,min=5,max=50,alpha,excludesall=!@#$%^&*,excludesall=0123456789"); err != nil {
	//	return fmt.Errorf("name must be between 5 and 50 characters and contain only letters")
	//}
	return nil
}
func (u *UserForm) validateRole(roleVal string) error {
	//if roleVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(roleVal, "min=1,max=10"); err != nil {
	//		return fmt.Errorf("role must be between 1 and 10 characters")
	//	}
	//}
	return nil
}
func (u *UserForm) validateBirth(birthVal string) error {
	//if birthVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(birthVal, "datetime=2006-01-02"); err != nil {
	//		return fmt.Errorf("birth must be in the format YYYY-MM-DD")
	//	}
	//}
	return nil
}
func (u *UserForm) validatePhone(phoneVal string) error {
	//if phoneVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(phoneVal, "min=10,max=15,numeric"); err != nil {
	//		return fmt.Errorf("phone must be between 10 and 15 characters and contain only numbers")
	//	}
	//}
	return nil
}
func (u *UserForm) validateDocument(docVal string) error {
	//if docVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(docVal, "min=10,max=15,numeric"); err != nil {
	//		return fmt.Errorf("document must be between 10 and 15 characters and contain only numbers")
	//	}
	//}
	return nil
}
func (u *UserForm) validateAddress(addressVal string) error {
	//if addressVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(addressVal, "min=5,max=150"); err != nil {
	//		return fmt.Errorf("address must be between 5 and 150 characters")
	//	}
	//}
	return nil
}
func (u *UserForm) validateCity(cityVal string) error {
	//if cityVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(cityVal, "min=2,max=50"); err != nil {
	//		return fmt.Errorf("city must be between 2 and 50 characters")
	//	}
	//}
	return nil
}
func (u *UserForm) validateState(stateVal string) error {
	//if stateVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(stateVal, "min=2,max=50"); err != nil {
	//		return fmt.Errorf("state must be between 2 and 50 characters")
	//	}
	//}
	return nil
}
func (u *UserForm) validateCountry(countryVal string) error {
	//if countryVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(countryVal, "min=2,max=50"); err != nil {
	//		return fmt.Errorf("country must be between 2 and 50 characters")
	//	}
	//}
	return nil
}
func (u *UserForm) validateZip(zipVal string) error {
	//if zipVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(zipVal, "min=5,max=10"); err != nil {
	//		return fmt.Errorf("zip code must be between 5 and 10 characters")
	//	}
	//}
	return nil
}
func (u *UserForm) validateURL(urlVal string) error {
	//if urlVal == "" {
	//	err := error(nil)
	//	return err
	//} else {
	//	if err := validate.Var(urlVal, "url,min=5,max=200"); err != nil {
	//		return fmt.Errorf("URL must be between 5 and 200 characters")
	//	}
	//}
	return nil
}

/*func createInput(ph, tp, val string, req bool, min, max int, err string, vld func(string) error) types.TuizInput {
	mu.Lock()
	input := types.TuizInput{
		Ph:  ph,
		Tp:  tp,
		Val: val,
		Req: req,
		Min: min,
		Max: max,
		Err: err,
		Vld: vld,
	}
	mu.Unlock()
	return input
}*/

func NewUserForm() *UserForm {
	validate = validator.New()
	return &UserForm{}
}
