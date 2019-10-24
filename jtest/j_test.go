package jtest

import (
	"io"
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("test error")
var errWrapped = errors.Wrap(io.ErrClosedPipe, "wrapping text")

func TestAssert(t *testing.T) {
	var errTest = errors.New("test error")

	tt := []struct {
		name             string
		expected, actual error
	}{
		{name: "nil"},
		{
			name:     "non-jettison",
			expected: io.EOF,
			actual:   io.EOF,
		},
		{
			name:     "jettison",
			expected: errTest,
			actual:   errTest,
		},
		{
			name:     "wrapped",
			expected: io.EOF,
			actual:   errors.Wrap(io.EOF, "wrapping text"),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			Assert(t, tc.expected, tc.actual)
		})
	}
}

func TestFailLog(t *testing.T) {
	expected := `No error in chain matches expected:
expected: EOF
actual:   io: read/write on closed pipe
`
	assert.Equal(t, expected, failLog(io.EOF, io.ErrClosedPipe))
}

func TestPretty(t *testing.T) {
	tt := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil",
			expected: "<nil>",
		},
		{
			name:     "non-jettison",
			err:      io.EOF,
			expected: "EOF",
		},
		{
			name:     "jettison",
			err:      errTest,
			expected: "test error\n- code: test error\n  message: test error\n  source: github.com/luno/jettison/jtest/j_test.go:11\n  parameters: []\n",
		},
		{
			name: "wrapped",
			err:  errWrapped,
			expected: `wrapping text: io: read/write on closed pipe
- code: wrapping text
  message: wrapping text
  source: github.com/luno/jettison/jtest/j_test.go:12
  parameters: []
- code: ""
  message: 'io: read/write on closed pipe'
  source: ""
  parameters: []
`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, pretty(tc.err))
		})
	}
}
