package roles

import (
	"fmt"
	c "github.com/faelmori/gkbxsrv/internal/models/abtract/commons"
	"gorm.io/gorm"
)

type TypeRepoRole struct{ *gorm.DB }

func (g *TypeRepoRole) Create(u *Role) (*Role, error) {
	err := g.DB.Create(u).Error
	if err != nil {
		return nil, fmt.Errorf("RoleImpl repository: failed to create RoleImpl: %w", err)
	}
	return u, nil
}
func (g *TypeRepoRole) FindOne(where ...interface{}) (*Role, error) {
	var u Role
	err := g.DB.Where(where[0], where[1:]).First(&u).Error // Use a pointer to u
	if err != nil {
		//return nil, logz.ErrorLog(fmt.Sprintf("RoleImpl repository: failed to find RoleImpl: %v", err), "GDBase")
		return nil, fmt.Errorf("RoleImpl repository: failed to find RoleImpl: %w", err)
	}
	return &u, nil // Return the pointer
}
func (g *TypeRepoRole) FindAll(where ...interface{}) ([]*Role, error) {
	var roles []*Role
	err := g.DB.Where(where[0], where[1:]).Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("RoleImpl repository: failed to find all roles: %w", err)
	}
	return roles, nil
}
func (g *TypeRepoRole) Update(u *Role) (*Role, error) { // Update by ID
	err := g.DB.Save(u).Error // Use Save to update all fields
	if err != nil {
		return nil, fmt.Errorf("RoleImpl repository: failed to update RoleImpl: %w", err)
	}
	return u, nil
}
func (g *TypeRepoRole) Delete(id uint) error { // Delete by ID
	err := g.DB.Delete(&RoleImpl{}, id).Error // Delete by ID
	if err != nil {
		return fmt.Errorf("RoleImpl repository: failed to delete RoleImpl: %w", err)
	}
	return nil
}
func (g *TypeRepoRole) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (g *TypeRepoRole) List(where ...interface{}) (*c.TableHandler, error) {
	var roles []Role
	err := g.DB.Where(where[0], where[1:]).Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("RoleImpl repository: failed to list roles: %w", err)
	}
	rows := make(map[int]map[string]string)
	for i, rl := range roles {
		if r, ok := rl.(Role); ok {
			rows[i] = map[string]string{
				"id":          r.GetID(),
				"name":        r.GetName(),
				"description": r.GetDescription(),
				"active":      fmt.Sprintf("%t", r.GetActive()),
			}
		}
	}
	return &c.TableHandler{Rows: rows}, nil
}

func NewRoleRepo(db *gorm.DB) RepoRole { return &TypeRepoRole{db} }
