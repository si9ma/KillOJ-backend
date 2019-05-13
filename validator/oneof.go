package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/go-playground/validator.v8"
)

func oneOf(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	vals := strings.Split(param, " ") // split by space

	var val string
	switch fieldKind {
	case reflect.String:
		val = field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val = strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val = strconv.FormatUint(field.Uint(), 10)
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}
	for i := 0; i < len(vals); i++ {
		if vals[i] == val {
			return true
		}
	}
	return false
}
