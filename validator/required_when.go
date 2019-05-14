package validator

import (
	"reflect"

	"github.com/si9ma/KillOJ-common/utils"

	"gopkg.in/go-playground/validator.v8"
)

func requireWhenFieldNotEmpty(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	f := reflect.Indirect(currentStructOrField).FieldByName(param)
	if !utils.IsZeroOfUnderlyingType(f.Interface()) {
		return !utils.IsZeroOfUnderlyingType(field.Interface())
	}

	return true
}
