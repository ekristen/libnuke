package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testSlice = []string{"alpha", "bravo", "charlie"}

func TestContains(t *testing.T) {
	cases := []struct {
		expected string
		result   bool
	}{
		{

			expected: "alpha",
			result:   true,
		},
		{
			expected: "delta",
			result:   false,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%t", tc.result), func(t *testing.T) {
			actual := StringSliceContains(testSlice, tc.expected)
			assert.Equal(t, tc.result, actual)
		})
	}
}
