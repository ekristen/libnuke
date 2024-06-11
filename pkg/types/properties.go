package types

import (
	"fmt"
	"reflect"
	"strings"
)

// Properties is a map of key-value pairs.
type Properties map[string]string

// NewProperties creates a new Properties map.
func NewProperties() Properties {
	props := make(Properties)
	props.SetTagPrefix("tag")
	return props
}

// NewPropertiesFromStruct creates a new Properties map from a struct.
func NewPropertiesFromStruct(data interface{}) Properties {
	return NewProperties().SetFromStruct(data)
}

func (p Properties) SetTagPrefix(prefix string) Properties {
	p["_tagPrefix"] = prefix
	return p
}

// String returns a string representation of the Properties map.
func (p Properties) String() string {
	var parts []string
	for k, v := range p {
		if strings.HasPrefix(k, "_") {
			continue
		}

		parts = append(parts, fmt.Sprintf(`%s: "%v"`, k, v))
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

// Get returns the value of a key in the Properties map.
func (p Properties) Get(key string) string {
	value, ok := p[key]
	if !ok {
		return ""
	}

	return value
}

// Set sets a key-value pair in the Properties map.
func (p Properties) Set(key string, value interface{}) Properties {
	if value == nil {
		return p
	}

	switch v := value.(type) {
	case *string:
		if v == nil {
			return p
		}
		p[key] = *v
	case []byte:
		p[key] = string(v)
	case *bool:
		if v == nil {
			return p
		}
		p[key] = fmt.Sprint(*v)
	case *int64:
		if v == nil {
			return p
		}
		p[key] = fmt.Sprint(*v)
	case *int:
		if v == nil {
			return p
		}
		p[key] = fmt.Sprint(*v)
	default:
		// Fallback to Stringer interface. This produces gibberish on pointers,
		// but is the only way to avoid reflection.
		p[key] = fmt.Sprint(value)
	}

	return p
}

// SetWithPrefix sets a key-value pair in the Properties map with a prefix.
func (p Properties) SetWithPrefix(prefix, key string, value interface{}) Properties {
	key = strings.TrimSpace(key)
	prefix = strings.TrimSpace(prefix)

	if key == "" {
		return p
	}

	if prefix != "" {
		key = fmt.Sprintf("%s:%s", prefix, key)
	}

	return p.Set(key, value)
}

// SetTag sets a tag key-value pair in the Properties map.
func (p Properties) SetTag(tagKey *string, tagValue interface{}) Properties {
	return p.SetTagWithPrefix("", tagKey, tagValue)
}

// SetTagWithPrefix sets a tag key-value pair in the Properties map with a prefix.
func (p Properties) SetTagWithPrefix(prefix string, tagKey *string, tagValue interface{}) Properties {
	if tagKey == nil {
		return p
	}

	keyStr := strings.TrimSpace(*tagKey)
	prefix = strings.TrimSpace(prefix)

	if keyStr == "" {
		return p
	}

	if prefix != "" {
		keyStr = fmt.Sprintf("%s:%s", prefix, keyStr)
	}

	keyStr = fmt.Sprintf("%s:%s", p.Get("_tagPrefix"), keyStr)

	return p.Set(keyStr, tagValue)
}

// Equals compares two Properties maps.
func (p Properties) Equals(o Properties) bool {
	if p == nil && o == nil {
		return true
	}

	if p == nil || o == nil {
		return false
	}

	if len(p) != len(o) {
		return false
	}

	for k, pv := range p {
		ov, ok := o[k]
		if !ok {
			return false
		}

		if pv != ov {
			return false
		}
	}

	return true
}

// SetFromStruct sets the Properties map from a struct by reading the structs fields
func (p Properties) SetFromStruct(data interface{}) Properties { //nolint:funlen,gocyclo
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() {
			continue
		}

		isSet := false

		switch value.Kind() {
		case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan:
			isSet = !value.IsNil()
		default:
			isSet = value.Interface() != reflect.Zero(value.Type()).Interface()
		}

		if !isSet {
			continue
		}

		propertyTag := field.Tag.Get("property")
		options := strings.Split(propertyTag, ",")
		name := field.Name
		prefix := ""
		tagPrefix := ""
		inline := false

		if options[0] == "-" {
			continue
		}

		if len(options) == 2 && options[1] == "inline" {
			inline = true
		}

		for _, option := range options {
			parts := strings.Split(option, "=")
			if len(parts) != 2 {
				continue
			}
			switch parts[0] {
			case "name":
				name = parts[1]
			case "prefix":
				prefix = parts[1]
			case "tagPrefix":
				tagPrefix = parts[1]
			}
		}

		if inline {
			p.SetFromStruct(value.Interface())
			continue
		}

		if tagPrefix != "" {
			p.SetTagPrefix(tagPrefix)
		}

		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		switch value.Kind() {
		case reflect.Struct:
			// do nothing
		case reflect.Map:
			for _, key := range value.MapKeys() {
				val := value.MapIndex(key)
				if key.Kind() == reflect.Ptr {
					key = key.Elem()
				}
				if val.Kind() == reflect.Ptr {
					val = val.Elem()
				}
				name = key.String()
				p.SetTagWithPrefix(prefix, &name, val.Interface())
			}
		case reflect.Slice:
			for j := 0; j < value.Len(); j++ {
				sliceValue := value.Index(j)
				if sliceValue.Kind() == reflect.Ptr {
					sliceValue = sliceValue.Elem()
				}
				if sliceValue.Kind() == reflect.Struct {
					sliceValueV := reflect.ValueOf(sliceValue.Interface())
					keyField := sliceValueV.FieldByName("Key")
					valueField := sliceValueV.FieldByName("Value")

					if keyField.Kind() == reflect.Ptr {
						keyField = keyField.Elem()
					}
					if valueField.Kind() == reflect.Ptr {
						valueField = valueField.Elem()
					}

					if keyField.IsValid() && valueField.IsValid() {
						p.SetTagWithPrefix(prefix, &[]string{keyField.Interface().(string)}[0], valueField.Interface())
					}
				}
			}
		default:
			p.SetWithPrefix(prefix, name, value.Interface())
		}
	}

	return p
}
