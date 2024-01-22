package utils_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

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
			_, _ = fmt.Fprintf(w, "%s\n", tc.want)
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
	_, _ = w.WriteString("a expected input \n")
	_ = w.Close()

	// Call the function
	err := utils.Prompt("expected input")
	assert.Error(t, err)
	assert.Equal(t, "aborted", err.Error())
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
