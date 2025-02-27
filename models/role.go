package models

import (
	"github.com/faelmori/gkbxsrv/internal/models"
	"gorm.io/gorm"
)

type RoleRepo struct{ models.RoleRepo }
type Role struct{ models.Role }

func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{models.NewRoleRepo(db)} }
func RoleFactory() *Role                { return &Role{models.RoleFactory()} }
