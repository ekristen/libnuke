package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var emptyCollection []Collection
var emptyMapping = map[string]string{}
var baseCollection = Collection{"ResourceA", "ResourceB", "ResourceC", "ResourceD", "ResourceE", "ResourceF"}

func TestResolveResourceTypes(t *testing.T) {
	cases := []struct {
		name         string
		base         Collection
		includes     []Collection
		excludes     []Collection
		alternatives []Collection
		mapping      map[string]string
		expected     Collection
	}{
		{
			name:         "empty",
			base:         Collection{},
			includes:     emptyCollection,
			excludes:     emptyCollection,
			alternatives: emptyCollection,
			mapping:      emptyMapping,
			expected:     Collection{},
		},
		{
			name:         "base",
			base:         baseCollection,
			includes:     emptyCollection,
			excludes:     emptyCollection,
			alternatives: emptyCollection,
			mapping:      emptyMapping,
			expected:     baseCollection,
		},
		{
			name:         "includes",
			base:         baseCollection,
			includes:     []Collection{{"ResourceA", "ResourceB", "ResourceC"}},
			excludes:     emptyCollection,
			alternatives: emptyCollection,
			mapping:      emptyMapping,
			expected:     Collection{"ResourceA", "ResourceB", "ResourceC"},
		},
		{
			name:         "excludes",
			base:         baseCollection,
			includes:     emptyCollection,
			excludes:     []Collection{{"ResourceA", "ResourceB", "ResourceC"}},
			alternatives: emptyCollection,
			mapping:      emptyMapping,
			expected:     Collection{"ResourceD", "ResourceE", "ResourceF"},
		},
		{
			name: "alternatives",
			base: Collection{
				"ResourceA",
				"ResourceB",
				"ResourceC",
				"ResourceD",
				"ResourceE",
				"ResourceF",
				"AlternativeA",
				"AlternativeC",
				"AlternativeE",
			},
			includes:     emptyCollection,
			excludes:     emptyCollection,
			alternatives: []Collection{{"AlternativeA", "AlternativeC", "AlternativeE"}},
			mapping: map[string]string{
				"AlternativeA": "ResourceA",
				"AlternativeC": "ResourceC",
				"AlternativeE": "ResourceE",
			},
			expected: Collection{"ResourceB", "ResourceD", "ResourceF", "AlternativeA", "AlternativeC", "AlternativeE"},
		},
		{
			name: "includes and excludes",
			base: baseCollection,
			includes: []Collection{
				{"ResourceA", "ResourceB", "ResourceC"},
			},
			excludes: []Collection{
				{"ResourceA", "ResourceB", "ResourceC"},
			},
			alternatives: emptyCollection,
			mapping:      emptyMapping,
			expected:     Collection{},
		},
		{
			name:     "excludes and alternatives",
			base:     baseCollection,
			includes: emptyCollection,
			excludes: []Collection{
				{"ResourceB", "ResourceC"},
			},
			alternatives: []Collection{
				{"AlternativeA", "AlternativeC", "AlternativeE"},
			},
			mapping: map[string]string{
				"AlternativeA": "ResourceA",
				"AlternativeC": "ResourceC",
				"AlternativeE": "ResourceE",
			},
			expected: Collection{"ResourceD", "ResourceF", "AlternativeA", "AlternativeC", "AlternativeE"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ResolveResourceTypes(tc.base, tc.includes, tc.excludes, tc.alternatives, tc.mapping)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
