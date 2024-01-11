package utils

import (
	"bytes"
	"math/rand"
	"time"
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
