package model_utils

import (
	"github.com/davidh16/goblin/utils"
	"reflect"
	"time"
)

var OptionalUserModelAttributes = map[string]reflect.Type{
	"Username":  utils.StringType,
	"FirstName": utils.StringType,
	"LastName":  utils.StringType,
	"Verified":  utils.BoolType,
	"LastLogin": utils.TimeType,
}

var AllPossibleUserModelAttributes = map[string]reflect.Type{
	"Uuid":      utils.UuidType,
	"Email":     utils.StringType,
	"Password":  utils.StringType,
	"CreatedAt": utils.TimeType,
	"UpdatedAt": utils.TimeType,
	"Username":  utils.StringType,
	"FirstName": utils.StringType,
	"LastName":  utils.StringType,
	"Verified":  utils.BoolType,
	"LastLogin": utils.TimeType,
}

var NonOptionalUserModelAttributeKeys = []string{"Uuid", "Email", "Password", "CreatedAt", "UpdatedAt"}

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
		DataType: OptionalUserModelAttributes[key].String(),
		JsonTag:  utils.GenerateJsonTag(key),
	}
}
