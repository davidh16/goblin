package user

import (
	"goblin/utils"
	"reflect"
	"time"
)

var optionalUserModelAttributes = map[string]reflect.Type{
	"Username":  utils.StringType,
	"FirstName": utils.StringType,
	"LastName":  utils.StringType,
	"Verified":  utils.BoolType,
	"LastLogin": utils.TimeType,
}

const (
	UserModelTemplatePath = "commands/model/flags/user/user.tmpl"
)

type UserModelAttribute struct {
	Label    string
	DataType string
	JsonTag  string
}

func NewUserModelAttribute(key string) UserModelAttribute {
	reflect.TypeOf(time.Now())
	return UserModelAttribute{
		Label:    key,
		DataType: optionalUserModelAttributes[key].String(),
		JsonTag:  utils.GenerateJsonTag(key),
	}
}
