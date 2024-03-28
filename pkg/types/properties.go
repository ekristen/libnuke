package types

import (
	"fmt"
	"reflect"
	"strings"
)

type Properties map[string]string

func NewProperties() Properties {
	return make(Properties)
}

func NewPropertiesFromStruct(data interface{}) Properties {
	return NewProperties().SetFromStruct(data)
}

func (p Properties) String() string {
	var parts []string
	for k, v := range p {
		parts = append(parts, fmt.Sprintf(`%s: "%v"`, k, v))
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

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

func (p Properties) SetTag(tagKey *string, tagValue interface{}) Properties {
	return p.SetTagWithPrefix("", tagKey, tagValue)
}

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

	keyStr = fmt.Sprintf("tag:%s", keyStr)

	return p.Set(keyStr, tagValue)
}

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

func (p Properties) Get(key string) string {
	value, ok := p[key]
	if !ok {
		return ""
	}

	return value
}

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

func (p Properties) SetFromStruct(data interface{}) Properties {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

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

		if options[0] == "-" {
			continue
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
			}
		}

		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		switch value.Kind() {
		case reflect.Map:
			for _, key := range value.MapKeys() {
				val := value.MapIndex(key)
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
		case reflect.Int:
			p.SetWithPrefix(prefix, field.Name, value.Interface().(int))
		case reflect.Int64:
			p.SetWithPrefix(prefix, field.Name, value.Interface().(int64))
		case reflect.String:
			p.SetWithPrefix(prefix, field.Name, value.Interface().(string))
		case reflect.Bool:
			p.SetWithPrefix(prefix, field.Name, value.Interface().(bool))
		default:
			panic(fmt.Errorf("unsupported type %v -> %v", value.Kind(), value.Interface()))
		}
	}

	return p
}

type KeyValue struct {
	Key   *string
	Value *string
}
