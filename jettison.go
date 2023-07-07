package jettison

import (
	"strings"
)

// Option allows one to attach metadata to an error or log.
type Option interface {
	Apply(Details)
}

// OptionFunc is a function-to-Option adapter.
type OptionFunc func(Details)

func (o OptionFunc) Apply(d Details) {
	o(d)
}

// Details provides methods to modify the metadata associated with an error or
// a log, such as arbitrary key/value pairs or stacktrace information.
type Details interface {
	SetKey(key, value string)
	SetSource(src string)
}

func WithKeyValueString(key, value string) OptionFunc {
	return func(d Details) {
		d.SetKey(normalise(key), value)
	}
}

func WithSource(src string) OptionFunc {
	return func(d Details) {
		d.SetSource(src)
	}
}

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
