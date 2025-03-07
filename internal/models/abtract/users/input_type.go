package users

type RegisterUserInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
type RegisterUserWithEmailInput struct {
	Username string `json:"username" binding:"omitempty"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
type RegisterUserWithUsernameInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"omitempty"`
	Password string `json:"password" binding:"required,min=8"`
}
type LoginWithEmailInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
type LoginWithUsernameInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}
