package unique

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
)

// Generate creates a unique hash from a slice of interface{} values.
func Generate(data ...interface{}) string {
	var allBytes []byte
	for _, d := range data {
		b, err := json.Marshal(d)
		if err != nil {
			// fallback: use fmt.Sprintf if marshal fails
			b = []byte(fmt.Sprintf("%v", d))
		}
		allBytes = append(allBytes, b...)
	}
	hash := sha256.Sum256(allBytes)
	return hex.EncodeToString(hash[:])
}

// FromStruct generates a unique key from a struct based on fields tagged with "libnuke:uniqueKey".
func FromStruct(data interface{}) *string {
	if data == nil {
		return nil
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	var values []interface{}
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

		values = append(values, valueField.Interface())
	}

	if len(values) == 0 {
		return nil
	}

	hash := Generate(values)
	return &hash
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
