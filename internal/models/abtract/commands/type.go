package commands

type Command interface {
	Name() string
	Validate(data interface{}) error
	Execute(data interface{}) (interface{}, error)
}
