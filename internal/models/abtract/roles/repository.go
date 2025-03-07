package roles

import c "github.com/faelmori/gkbxsrv/internal/models/abtract/commons"

type RepoRole interface {
	Create(u *Role) (*Role, error)
	FindOne(where ...interface{}) (*Role, error)
	FindAll(where ...interface{}) ([]*Role, error)
	Update(u *Role) (*Role, error)
	Delete(id uint) error
	Close() error
	List(where ...interface{}) (*c.TableHandler, error)
}
