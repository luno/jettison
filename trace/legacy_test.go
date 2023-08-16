package trace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitiseFnName(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{
		{},
		{"main.main", "main"},
		{"foo/bar.baz", "baz"},
		{"foo/bar.(*Baz)·qux", "(*Baz).qux"},
		{"foo.bar/baz.qux", "qux"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.in, func(t *testing.T) {
			out := sanitiseFnName(testCase.in)
			require.Equal(t, testCase.out, out)
		})
	}
}
