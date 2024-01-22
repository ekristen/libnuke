// Package types provides common types used by libnuke. Primarily it provides the Collection type which is used to
// represent a collection of strings. Additionally, it provides the Properties type which is used to add properties
// to a resource.
package types

// Collection is a collection of strings
type Collection []string

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
