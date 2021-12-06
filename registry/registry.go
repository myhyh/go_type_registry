package registry

import (
	"fmt"
	"reflect"
)


var Registry = map[interface{}]string{}
var RevRegistry = map[string]interface{}{}

// 根据字符串得到
func New(s string)interface{}{
	t := RevRegistry[s]
	if t != nil {
		return reflect.New(reflect.TypeOf(t).Elem())
	}
	return fmt.Errorf("type %s not found",s)
}

func GetTypeString(t interface{}) string{
	return reflect.ValueOf(t).Type().PkgPath()+"/"+reflect.ValueOf(t).Type().Name()
}

func GetRegisteredTypeName(t interface{}) string{
	return Registry[t]
}

func Register(i interface{}) bool{
	s := GetTypeString(i)
	Registry[i]=s
	RevRegistry[s]=i
	return true
}