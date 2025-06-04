package unique

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
)

func FromStruct(data interface{}) (*string, error) {
	if data == nil {
		return nil, nil
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, nil
	}

	var values []string
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := field.Tag.Get("libnuke")
		if tag != "uniqueKey" {
			continue
		}

		valueField := v.Field(i)
		if !valueField.IsValid() || (valueField.Kind() == reflect.Slice && valueField.Len() == 0) {
			continue
		}

		value := valueField.Interface()
		values = append(values, toString(value))
	}

	if len(values) == 0 {
		return nil, nil
	}

	combined := strings.Join(values, ",")
	hash := sha256.Sum256([]byte(combined))
	result := hex.EncodeToString(hash[:])
	return &result, nil
}

// toString converts interface{} to string for basic types
func toString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return reflect.ValueOf(v).String()
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}
