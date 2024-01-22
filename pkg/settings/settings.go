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

func (s *Setting) Get(key string) interface{} {
	value, ok := (*s)[key]
	if !ok {
		return nil
	}

	switch value.(type) {
	case string:
		return value.(string)
	case int:
		return value.(int)
	case bool:
		return value.(bool)
	default:
		return value
	}
}

func (s *Setting) Set(key string, value interface{}) {
	(*s)[key] = value
}
