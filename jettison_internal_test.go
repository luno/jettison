package jettison

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalise(t *testing.T) {
	testCases := []struct {
		in  string
		exp string
	}{
		{in: "lowercase", exp: "lowercase"},
		{in: "UPPERCASE", exp: "uppercase"},
		{in: "numbers0123456789", exp: "numbers0123456789"},
		{in: "special-_.", exp: "special-_."},
		{in: "grpc-prefix", exp: "prefix"},
		{in: "disallowed !@#$%^&*()'", exp: "disallowed"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.exp, normalise(tc.in))
		})
	}
}
