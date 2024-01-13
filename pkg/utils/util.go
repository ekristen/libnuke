// Package utils provides several helper functions used throughout the library or useful to the upstream tools
// that implement the primary parts of the library
package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ekristen/libnuke/pkg/types"
)

const Base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const UniqueIDLength = 6 // Should be good for 62^6 = 56+ billion combinations

// UniqueID - Returns a unique (ish) id we can attach to resources and tfstate files, so they don't conflict
// with each other. Uses base 62 to generate a 6 character string that's unlikely to collide with the handful
// of tests we run in parallel. Based on code here: http://stackoverflow.com/a/9543797/483528
func UniqueID() string {
	var out bytes.Buffer

	randVal := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint: gosec
	for i := 0; i < UniqueIDLength; i++ {
		out.WriteByte(Base62Chars[randVal.Intn(len(Base62Chars))])
	}

	return out.String()
}

// Prompt creates a prompt for direct user interaction to receive input
func Prompt(expect string) error {
	fmt.Print("> ")
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	if strings.TrimSpace(text) != expect {
		return fmt.Errorf("aborted")
	}
	fmt.Println()

	return nil
}

func ResolveResourceTypes(base types.Collection, include, exclude []types.Collection) types.Collection {
	for _, i := range include {
		if len(i) > 0 {
			base = base.Intersect(i)
		}
	}

	for _, e := range exclude {
		base = base.Remove(e)
	}

	return base
}

func IsTrue(s string) bool {
	return strings.TrimSpace(strings.ToLower(s)) == "true"
}
