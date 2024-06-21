// Package settings provides a way to store and retrieve settings for resources.
package settings

type Settings map[string]*Setting

func (s *Settings) Get(key string) *Setting {
	if s == nil {
		return nil
	}

	set, ok := (*s)[key]
	if !ok {
		return nil
	}

	return set
}

func (s *Settings) Set(key string, value *Setting) {
	existing, ok := (*s)[key]
	if ok {
		for k, v := range *value {
			(*existing)[k] = v
		}

		return
	}

	(*s)[key] = value
}

type Setting map[string]interface{}

// Get returns the value of a key in the Setting
// Deprecated: use GetBool, GetString, or GetInt instead
func (s *Setting) Get(key string) interface{} {
	value, ok := (*s)[key]
	if !ok {
		return nil
	}

	switch value.(type) {
	case string:
		return value
	case int:
		return value
	case bool:
		return value
	default:
		return value
	}
}

// GetBool returns the boolean value of a key in the Setting
func (s *Setting) GetBool(key string) bool {
	value, ok := (*s)[key]
	if !ok {
		return false
	}

	return value.(bool)
}

// GetString returns the string value of a key in the Setting
func (s *Setting) GetString(key string) string {
	value, ok := (*s)[key]
	if !ok {
		return ""
	}

	return value.(string)
}

// GetInt returns the integer value of a key in the Setting
func (s *Setting) GetInt(key string) int {
	value, ok := (*s)[key]
	if !ok {
		return 0
	}

	return value.(int)
}

// Set sets a key value pair in the Setting
func (s *Setting) Set(key string, value interface{}) {
	(*s)[key] = value
}
