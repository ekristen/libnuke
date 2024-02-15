// Package filter provides a way to filter resources based on a set of criteria.
package filter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mb0/glob"
)

type OpType string
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

	And OpType = "and"
	Or  OpType = "or"

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

// Append appends the filters from f2 to f
func (f Filters) Append(f2 Filters) {
	for resourceType, filter := range f2 {
		f[resourceType] = append(f[resourceType], filter...)
	}
}

// Merge is an alias of Append for backwards compatibility
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

	m := map[string]string{}
	err := unmarshal(m)
	if err != nil {
		fmt.Println("%%%%%%%%")
		return err
	}

	f.Type = Type(m["type"])
	f.Value = m["value"]
	f.Property = m["property"]
	f.Invert = m["invert"]
	return nil
}

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
