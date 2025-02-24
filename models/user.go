package models

import (
	"github.com/faelmori/gokubexfs/internal/models"
	"gorm.io/gorm"
)

type UserRepo struct{ models.UserRepo }
type User struct{ models.User }

func NewUserRepo(db *gorm.DB) *UserRepo         { return &UserRepo{models.NewUserRepo(db)} }
func UserFactory(userData map[string]any) *User { return &User{models.UserFactory(userData)} }
