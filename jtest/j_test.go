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
	t.Run("log without message", func(t *testing.T) {
		expected := `No error in chain matches expected:
expected: EOF
actual:   io: read/write on closed pipe
`
		assert.Equal(t, expected, failLog(io.EOF, io.ErrClosedPipe))
	})

	t.Run("log with message", func(t *testing.T) {
		expected := `No error in chain matches expected:
expected: EOF
actual:   io: read/write on closed pipe
message:  errors in chain check
`
		assert.Equal(t, expected, failLog(io.EOF, io.ErrClosedPipe, "errors in chain check"))
	})
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

func Test_messageFromMsg(t *testing.T) {
	type args struct {
		msg []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "without message",
			args: args{
				msg: makeInterfaceSlice(),
			},

			want: "",
		},
		{
			name: "with message",
			args: args{
				msg: makeInterfaceSlice("check the message"),
			},

			want: "check the message",
		},
		{
			name: "with non a string message",
			args: args{
				msg: makeInterfaceSlice(42),
			},

			want: "42",
		},
		{
			name: "with more than one argument",
			args: args{
				msg: makeInterfaceSlice("first argument", 42),
			},

			want: "first argument 42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := messageFromMsgs(tt.args.msg...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func makeInterfaceSlice(al ...interface{}) []interface{} {
	return al
}
