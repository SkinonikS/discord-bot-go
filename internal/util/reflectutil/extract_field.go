package reflectutil

import (
	"reflect"
)

func ExtractField(obj any, fieldName string) (any, bool) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, false
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, false
	}

	return field.Interface(), true
}
