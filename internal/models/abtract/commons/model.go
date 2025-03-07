package commons

type Model interface {
	TableName() string
	GetID() string
	Validate() error
}
