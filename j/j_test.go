package j

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luno/jettison/errors"
)

// fmtonly tests sprint if fmt.Formatter but not fmt.Stringer.
type fmtonly struct{}

func (_ fmtonly) Format(s fmt.State, c rune) {
	s.Write([]byte("fmtonly"))
}

func TestSprint(t *testing.T) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.Output, sprint(test.Input))
		})
	}
}

// BenchmarkSimple benchmarks sprinting all values types.
func BenchmarkAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			sprint(test.Input)
		}
	}
}

// BenchmarkAllOld benchmarks fmt.Sprint all values types.
func BenchmarkAllOld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			_ = fmt.Sprint(test.Input)
		}
	}
}

// BenchmarkSimple benchmarks sprinting simple values (strings, numbers, bools).
func BenchmarkSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, input := range simple {
			sprint(input)
		}
	}
}

// BenchmarkSimpleOld benchmarks fmt.Sprint simple values (strings, numbers, bools).
func BenchmarkSimpleOld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, input := range simple {
			_ = fmt.Sprint(input)
		}
	}
}

// BenchmarkFull runs the sprint and fmt.Sprint functions against all types in "tests"
func BenchmarkFull(b *testing.B) {
	for _, t := range tests {
		b.Run(t.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = sprint(t.Input)
			}
		})
		b.Run(t.Name+"Fmt", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = fmt.Sprint(t.Input)
			}
		})
	}
}

var simple = []interface{}{1, 1.0, "1", true}

var tests = []struct {
	Name   string
	Input  interface{}
	Output string
}{
	{
		Name:   "nil",
		Input:  nil,
		Output: "<nil>",
	}, {
		Name:   "nil stringer",
		Input:  (fmt.Stringer)(nil),
		Output: "<nil>",
	}, {
		Name:   "bool",
		Input:  true,
		Output: "true",
	}, {
		Name:   "int",
		Input:  1,
		Output: "1",
	}, {
		Name:   "int64",
		Input:  int64(64),
		Output: "64",
	}, {
		Name:   "uint32",
		Input:  uint32(32),
		Output: "32",
	}, {
		Name:   "float64",
		Input:  1.0,
		Output: "1",
	}, {
		Name:   "string",
		Input:  "string",
		Output: "string",
	}, {
		Name:   "stringer enum",
		Input:  reflect.Interface,
		Output: "interface",
	}, {
		Name:   "stringer struct",
		Input:  time.Time{},
		Output: "0001-01-01 00:00:00 +0000 UTC",
	}, {
		Name:   "stringer struct pointer",
		Input:  new(time.Time),
		Output: "0001-01-01 00:00:00 +0000 UTC",
	}, {
		Name:   "struct pointer",
		Input:  new(sync.Mutex),
		Output: "<ptr>",
	}, {
		Name:   "struct",
		Input:  sync.Mutex{},
		Output: "<struct>",
	}, {
		Name:   "chan",
		Input:  make(chan struct{}, 0),
		Output: "<chan>",
	}, {
		Name:   "array",
		Input:  [2]int{1, 2},
		Output: "<array>",
	}, {
		Name:   "slice",
		Input:  []int{1, 2},
		Output: "<slice>",
	}, {
		Name:   "formatter only",
		Input:  fmtonly{},
		Output: "fmtonly",
	}, {
		Name:   "func",
		Input:  func() {},
		Output: "<func>",
	},
}

func TestC(t *testing.T) {
	errFoo := errors.New("foo", C("123"))

	je := errFoo.(*errors.JettisonError)
	require.Empty(t, je.Hops[0].StackTrace)
	require.Equal(t, "123", je.Hops[0].Errors[0].Code)

	err := errors.Wrap(errFoo, "wrap adds stacktrace")
	je = err.(*errors.JettisonError)
	require.NotEmpty(t, je.Hops[0].StackTrace)
	require.True(t, errors.Is(err, errFoo))
}

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
