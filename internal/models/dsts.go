package models

import "reflect"

var ModelList = []interface{}{
	&UserImpl{},
	&productImpl{},
	&CustomerImpl{},
	&OrderImpl{},
}

var ModelRegistry = map[string]reflect.Type{
	"User":     reflect.TypeOf(UserImpl{}),
	"Product":  reflect.TypeOf(productImpl{}),
	"Customer": reflect.TypeOf(CustomerImpl{}),
	"Order":    reflect.TypeOf(OrderImpl{}),
}
