package utils_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"sort"
	"testing"

	"github.com/ekristen/libnuke/pkg/types"
	"github.com/ekristen/libnuke/pkg/utils"
)

func TestUniqueID(t *testing.T) {
	id := utils.UniqueID()
	assert.Len(t, id, utils.UniqueIDLength)
}

func TestPrompt(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{
			name: "simple",
			want: "simple",
		},
		{
			name: "with-spaces",
			want: "simple prompt",
		},
		{
			name: "with\ttabs",
			want: "with\ttabs",
		},
		{
			name: "with-special-chars",
			want: "another prompt with $ chars #",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a pipe
			r, w, _ := os.Pipe()

			// Replace the standard input with our pipe
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }() // Restore original Stdin
			os.Stdin = r

			// Write our input into the pipe
			_, _ = w.Write([]byte(fmt.Sprintf("%s\n", tc.want)))
			_ = w.Close()

			// Call the function
			err := utils.Prompt(tc.want)

			// Check the result
			if err != nil {
				t.Errorf("Prompt returned an error: %v", err)
			}
		})
	}
}

func TestPromptError(t *testing.T) {
	// Create a pipe
	r, w, _ := os.Pipe()

	// Replace the standard input with our pipe
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin
	os.Stdin = r

	// Close the write end of the pipe to simulate an error
	_ = w.Close()

	// Call the function
	err := utils.Prompt("expected input")
	assert.Error(t, err)
}

func TestPromptTrimSpace(t *testing.T) {
	// Create a pipe
	r, w, _ := os.Pipe()

	// Replace the standard input with our pipe
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin
	os.Stdin = r

	// Write our input into the pipe
	_, _ = w.Write([]byte("a expected input \n"))
	_ = w.Close()

	// Call the function
	err := utils.Prompt("expected input")
	assert.Error(t, err)
	assert.Equal(t, "aborted", err.Error())
}

func TestResolveResourceTypes(t *testing.T) {
	cases := []struct {
		base    types.Collection
		include []types.Collection
		exclude []types.Collection
		result  types.Collection
	}{
		{
			base:    types.Collection{"a", "b", "c", "d"},
			include: []types.Collection{{"a", "b", "c"}},
			result:  types.Collection{"a", "b", "c"},
		},
		{
			base:    types.Collection{"a", "b", "c", "d"},
			exclude: []types.Collection{{"b", "d"}},
			result:  types.Collection{"a", "c"},
		},
		{
			base:    types.Collection{"a", "b"},
			include: []types.Collection{{}},
			result:  types.Collection{"a", "b"},
		},
		{
			base:    types.Collection{"c", "b"},
			exclude: []types.Collection{{}},
			result:  types.Collection{"c", "b"},
		},
		{
			base:    types.Collection{"a", "b", "c", "d"},
			include: []types.Collection{{"a", "b", "c"}},
			exclude: []types.Collection{{"a"}},
			result:  types.Collection{"b", "c"},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			r := utils.ResolveResourceTypes(tc.base, tc.include, tc.exclude)

			sort.Strings(r)
			sort.Strings(tc.result)

			var (
				want = fmt.Sprint(tc.result)
				have = fmt.Sprint(r)
			)

			if want != have {
				t.Fatalf("Wrong result. Want: %s. Have: %s", want, have)
			}
		})
	}
}

func TestIsTrue(t *testing.T) {
	falseStrings := []string{"", "false", "treu", "foo"}
	for _, fs := range falseStrings {
		if utils.IsTrue(fs) {
			t.Fatalf("IsTrue falsely returned 'true' for: %s", fs)
		}
	}

	trueStrings := []string{"true", " true", "true ", " TrUe "}
	for _, ts := range trueStrings {
		if !utils.IsTrue(ts) {
			t.Fatalf("IsTrue falsely returned 'false' for: %s", ts)
		}
	}
}
