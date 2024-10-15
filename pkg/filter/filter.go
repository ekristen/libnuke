// Package filter provides a way to filter resources based on a set of criteria.
package filter

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
	"slices"
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
	NotIn         Type = "NotIn"
	In            Type = "In"

	And OpType = "and"
	Or  OpType = "or"

	Global = "__global__"
)

type Property interface {
	GetProperty(string) (string, error)
}

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

// GetByGroup returns the filters grouped by the group name for a specific resource type. If there are no filters it
// returns nil
func (f Filters) GetByGroup(resourceType string) map[string][]Filter {
	filters := make(Filters)

	if f[resourceType] != nil {
		for _, filter := range f[resourceType] {
			group := filter.GetGroup()
			if filters[group] == nil {
				filters[group] = []Filter{}
			}

			filters[group] = append(filters[group], filter)
		}
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

// Match checks if the filters match the given property which is actually a queue item that meats the
// property interface requirements
func (f Filters) Match(resourceType string, p Property) (bool, error) {
	resourceFilters := f.GetByGroup(resourceType)
	if resourceFilters == nil {
		return false, nil
	}

	groupCount := make(map[string]int)
	totalCount := make(map[string]int)
	matchCount := make(map[string]int)

	for group, groupFilters := range resourceFilters {
		totalCount[group] = len(groupFilters)

		for _, filter := range groupFilters {
			prop, err := p.GetProperty(filter.Property)
			if err != nil {
				logrus.WithError(err).Warn("error getting property")
				return false, err
			}

			match, err := filter.Match(prop)
			if err != nil {
				logrus.WithError(err).Warn("error matching filter")
				return false, err
			}

			logrus.Trace(match, filter.Invert)

			if filter.Invert {
				match = !match
			}

			if match {
				matchCount[group]++
			}
		}

		logrus.Trace("groupCount", groupCount)
		logrus.Trace("totalCount", totalCount)
		logrus.Trace("matchCount", matchCount)

		if totalCount[group] == matchCount[group] {
			groupCount[group]++
		}
	}

	// If the group count is greater than 0, then one of the groups matched
	if len(groupCount) > 0 {
		return true, nil
	}

	return false, nil
}

// Filter is a filter to apply to a resource
type Filter struct {
	// Group is the name of the group of filters, all filters in a group are ANDed together
	Group string `yaml:"group" json:"group"`

	// Type is the type of filter to apply
	Type Type `yaml:"type" json:"type"`

	// Property is the name of the property to filter on
	Property string `yaml:"property" json:"property"`

	// Value is the value to filter on
	Value string `yaml:"value" json:"value"`

	// Values allows for multiple values to be specified for a filter
	Values []string `yaml:"values" json:"values"`

	// Invert is a flag to invert the filter
	Invert bool `yaml:"invert" json:"invert"`
}

// GetGroup returns the group name of the filter, if it is empty it returns "default"
func (f *Filter) GetGroup() string {
	if f.Group == "" {
		return "default"
	}
	return f.Group
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

// UnmarshalYAML unmarshals a filter from YAML data
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
		f.Invert = false
	} else {
		switch m["invert"].(type) {
		case bool:
			f.Invert = m["invert"].(bool)
		case string:
			invert, err := strconv.ParseBool(m["invert"].(string))
			if err != nil {
				return err
			}
			f.Invert = invert
		}
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
