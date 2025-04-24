package jtest

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

func TestAssert(t *testing.T) {
	errTest := errors.New("test error")

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

func TestRequire(t *testing.T) {
	errTest := errors.New("test error")

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
			Require(t, tc.expected, tc.actual)
		})
	}
}

func TestAssertNil(t *testing.T) {
	AssertNil(t, nil)
}

func TestRequireNil(t *testing.T) {
	RequireNil(t, nil)
}

func TestFailLog(t *testing.T) {
	t.Run("log without message", func(t *testing.T) {
		expected := `No error in chain matches expected:
expected: - message: EOF
actual:   - message: 'io: read/write on closed pipe'
`
		require.Equal(t, expected, failLog(io.EOF, io.ErrClosedPipe))
	})

	t.Run("log with message", func(t *testing.T) {
		expected := `No error in chain matches expected:
expected: - message: EOF
actual:   - message: 'io: read/write on closed pipe'
message:  errors in chain check
`
		require.Equal(t, expected, failLog(io.EOF, io.ErrClosedPipe, "errors in chain check"))
	})
}

func TestFailNilLog(t *testing.T) {
	t.Run("log without message", func(t *testing.T) {
		expected := `Unexpected non-nil error:
actual:   - message: 'io: read/write on closed pipe'
`
		require.Equal(t, expected, failNilLog(io.ErrClosedPipe))
	})

	t.Run("log with message", func(t *testing.T) {
		expected := `Unexpected non-nil error:
actual:   - message: EOF
message:  errors in chain check
`
		require.Equal(t, expected, failNilLog(io.EOF, "errors in chain check"))
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
			expected: "- message: EOF\n",
		},
		{
			name:     "jettison",
			err:      errors.New("test error", j.C("ERR_48026e342952be11")),
			expected: "- message: test error\n  code: ERR_48026e342952be11\n",
		},
		{
			name:     "wrapped",
			err:      errors.Wrap(io.ErrClosedPipe, "wrapping text"),
			expected: "- message: 'wrapping text: io: read/write on closed pipe'\n",
		},
		{
			name: "joined",
			err: errors.Wrap(errors.Join(
				errors.New("a", j.C("error_a")),
				errors.New("b", j.C("error_b")),
			), "wrap", j.KV("wrap", "true")),
			expected: `- message: a
  code: error_a
  kv:
    - key: wrap
      value: "true"
- message: b
  code: error_b
  kv:
    - key: wrap
      value: "true"
`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, pretty(tc.err))
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

			want: "message:  check the message\n",
		},
		{
			name: "with non-string message",
			args: args{
				msg: makeInterfaceSlice(42),
			},

			want: "message:  42\n",
		},
		{
			name: "with more than one argument",
			args: args{
				msg: makeInterfaceSlice("first argument", 42),
			},

			want: "message:  first argument 42\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := messageFromMsgs(tt.args.msg...)
			require.Equal(t, tt.want, got)
		})
	}
}

func makeInterfaceSlice(al ...interface{}) []interface{} {
	return al
}
