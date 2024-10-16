package docs

import (
	"fmt"
	"reflect"
	"strings"
)

func GeneratePropertiesMap(data interface{}) map[string]string {
	properties := map[string]string{}

	if data == nil {
		return properties
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
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

		if prefix != "" && name != "Tags" {
			name = fmt.Sprintf("%s:%s", prefix, name)
		}

		descriptionTag := field.Tag.Get("description")

		if name == "Tags" {
			originalName := name
			name = "tag:<key>:"
			tagPrefix := "tag:"
			if prefix != "" {
				tagPrefix = fmt.Sprintf("tag:%s:", prefix)
			}

			descriptionTag = fmt.Sprintf(
				"This resource has tags with property `%s`. These are key/value pairs that are\n\t"+
					"added as their own property with the prefix of `%s` (e.g. [%sexample: \"value\"]) ",
				originalName, tagPrefix, tagPrefix)

			if prefix != "" {
				name = fmt.Sprintf("tag:%s:<key>:", prefix)
			}
		}

		properties[name] = descriptionTag
	}

	return properties
}
