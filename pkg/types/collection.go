// Package types provides common types used by libnuke. Primarily it provides the Collection type which is used to
// represent a collection of strings. Additionally, it provides the Properties type which is used to add properties
// to a resource.
package types

import (
	"github.com/mb0/glob"
)

// Collection is a collection of strings
type Collection []string

// Expand returns a collection by using the Collection which may contain glob patterns and match to the source
// and returns the expanded collection, if there are no matches, it includes the original element from the collection.
func (c Collection) Expand(base []string) Collection {
	var expanded Collection
	for _, sc := range c {
		matches, _ := glob.GlobStrings(base, sc)

		if matches == nil {
			expanded = append(expanded, sc)
			continue
		}

		expanded = append(expanded, matches...)
	}

	return expanded
}

// Intersect returns the intersection of two collections
func (c Collection) Intersect(o Collection) Collection {
	mo := o.toMap()

	result := Collection{}
	for _, t := range c {
		if mo[t] {
			result = append(result, t)
		}
	}

	return result
}

// Remove returns the difference of two collections
func (c Collection) Remove(o Collection) Collection {
	mo := o.toMap()

	result := Collection{}
	for _, t := range c {
		if !mo[t] {
			result = append(result, t)
		}
	}

	return result
}

// Union returns the union of two collections
func (c Collection) Union(o Collection) Collection {
	ms := c.toMap()

	result := []string(c)
	for _, oi := range o {
		if !ms[oi] {
			result = append(result, oi)
		}
	}

	return result
}

// toMap converts a collection to a map
func (c Collection) toMap() map[string]bool {
	m := map[string]bool{}
	for _, t := range c {
		m[t] = true
	}
	return m
}
