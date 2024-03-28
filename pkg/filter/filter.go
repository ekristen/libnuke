// Package filter provides a way to filter resources based on a set of criteria.
package filter

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mb0/glob"
)

type Type string

const (
	Empty         Type = ""
	Exact         Type = "exact"
	Glob          Type = "glob"
	Regex         Type = "regex"
	Contains      Type = "contains"
	DateOlderThan Type = "dateOlderThan"
	Suffix        Type = "suffix"
	Prefix        Type = "prefix"
	NotIn         Type = "NotIn"
	In            Type = "In"

	Global = "__global__"
)

type Filters map[string][]Filter

// Get returns the filters for a specific resource type or the global filters if they exist. If there are no filters it
// returns nil
func (f Filters) Get(resourceType string) []Filter {
	var filters []Filter

	if f[Global] != nil {
		filters = append(filters, f[Global]...)
	}

	if f[resourceType] != nil {
		filters = append(filters, f[resourceType]...)
	}

	if len(filters) == 0 {
		return nil
	}

	return filters
}

// Validate checks if the filters are valid or not and returns an error if they are not
func (f Filters) Validate() error {
	for resourceType, filters := range f {
		for _, filter := range filters {
			if err := filter.Validate(); err != nil {
				return fmt.Errorf("%s: has an invalid filter: %+v", resourceType, filter)
			}
		}
	}

	return nil
}

// Append appends the filters from f2 to f. This is primarily used to append filters from a preset
// to a set of filters that were defined on a resource type.
func (f Filters) Append(f2 Filters) {
	for resourceType, filter := range f2 {
		f[resourceType] = append(f[resourceType], filter...)
	}
}

// Merge is an alias of Append for backwards compatibility
// Deprecated: use Append instead
func (f Filters) Merge(f2 Filters) {
	f.Append(f2)
}

// Filter is a filter to apply to a resource
type Filter struct {
	// Type is the type of filter to apply
	Type Type

	// Property is the name of the property to filter on
	Property string

	// Value is the value to filter on
	Value string

	// Values allows for multiple values to be specified for a filter
	Values []string

	// Invert is a flag to invert the filter
	Invert string
}

// Validate checks if the filter is valid
func (f *Filter) Validate() error {
	if f.Property == "" && f.Value == "" {
		return fmt.Errorf("property and value cannot be empty")
	}

	return nil
}

// Match checks if the filter matches the given value
func (f *Filter) Match(o string) (bool, error) {
	switch f.Type {
	case Empty, Exact:
		return f.Value == o, nil

	case Contains:
		return strings.Contains(o, f.Value), nil

	case Glob:
		return glob.Match(f.Value, o)

	case Regex:
		re, err := regexp.Compile(f.Value)
		if err != nil {
			return false, err
		}
		return re.MatchString(o), nil

	case DateOlderThan:
		if o == "" {
			return false, nil
		}
		duration, err := time.ParseDuration(f.Value)
		if err != nil {
			return false, err
		}
		fieldTime, err := parseDate(o)
		if err != nil {
			return false, err
		}
		fieldTimeWithOffset := fieldTime.Add(duration)

		return fieldTimeWithOffset.After(time.Now()), nil

	case Prefix:
		return strings.HasPrefix(o, f.Value), nil

	case Suffix:
		return strings.HasSuffix(o, f.Value), nil

	case In:
		return slices.Contains(f.Values, o), nil

	case NotIn:
		return !slices.Contains(f.Values, o), nil

	default:
		return false, fmt.Errorf("unknown type %s", f.Type)
	}
}

func (f *Filter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string

	if unmarshal(&value) == nil {
		f.Type = Exact
		f.Value = value
		return nil
	}

	m := map[string]interface{}{}
	err := unmarshal(m)
	if err != nil {
		fmt.Println("%%%%%%%%")
		return err
	}

	if m["type"] == nil {
		f.Type = Exact
	} else {
		f.Type = Type(m["type"].(string))
	}

	if m["value"] == nil {
		f.Value = ""
	} else {
		f.Value = m["value"].(string)
	}

	if m["values"] == nil {
		f.Values = []string{}
	} else {
		interfaceSlice := m["values"].([]interface{})
		stringSlice := make([]string, len(interfaceSlice))
		for i, v := range interfaceSlice {
			str, _ := v.(string)
			stringSlice[i] = str
		}

		f.Values = stringSlice
	}

	if m["property"] == nil {
		f.Property = ""
	} else {
		f.Property = m["property"].(string)
	}

	if m["invert"] == nil {
		f.Invert = ""
	} else {
		f.Invert = m["invert"].(string)
	}

	return nil
}

// NewExactFilter creates a new filter that matches the exact value
func NewExactFilter(value string) Filter {
	return Filter{
		Type:  Exact,
		Value: value,
	}
}

// parseDate parses a date from a string, it supports unix timestamps and RFC3339 formatted dates
func parseDate(input string) (time.Time, error) {
	if i, err := strconv.ParseInt(input, 10, 64); err == nil {
		t := time.Unix(i, 0)
		return t, nil
	}

	formats := []string{"2006-01-02",
		"2006/01/02",
		"2006-01-02T15:04:05Z",
		time.RFC3339Nano, // Format of t.MarshalText() and t.MarshalJSON()
		time.RFC3339,
	}
	for _, f := range formats {
		t, err := time.Parse(f, input)
		if err == nil {
			return t, nil
		}
	}
	return time.Now(), fmt.Errorf("unable to parse time %s", input)
}
