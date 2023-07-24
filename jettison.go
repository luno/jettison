package jettison

import (
	"strings"
)

var (
	allowedChars    = "0123456789abcdefghijklmnopqrstuvwxyz-_."
	allowedCharsMap map[rune]bool
)

func init() {
	allowedCharsMap = make(map[rune]bool)
	for _, ch := range allowedChars {
		allowedCharsMap[ch] = true
	}
}

// normalise modifies the given key to conform to gRPC metadata requirements,
// as the keys have to be transmittable over the wire (in contexts, for
// instance).
// See https://godoc.org/google.golang.org/grpc/metadata#New.
func normalise(key string) string {
	// Uppercase characters are normalised to lower case.
	key = strings.ToLower(key)

	// Keys beginning with 'grpc-' are disallowed.
	key = strings.TrimPrefix(key, "grpc-")

	var res string
	for _, ch := range key {
		// Remove illegal characters from the key.
		if !allowedCharsMap[ch] {
			continue
		}

		res += string(ch)
	}

	return res
}
