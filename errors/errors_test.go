package errors_test

import (
	stdlib_errors "errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/luno/jettison"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name string
		msg  string
		opts []jettison.Option
	}{
		{
			name: "key/value setting",
			msg:  "key/value setting",
			opts: []jettison.Option{
				jettison.WithKeyValueString("key", "value"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			line := nextLine()
			err := errors.New(tc.msg, tc.opts...)

			je, ok := err.(*errors.JettisonError)
			require.True(t, ok)

			assert.Len(t, je.Hops, 1)
			assert.Len(t, je.Hops[0].Errors, 1)
			assert.Equal(t, tc.msg, je.Hops[0].Errors[0].Message)
			assert.Equal(t,
				"github.com/luno/jettison/errors/errors_test.go:"+line,
				je.Hops[0].Errors[0].Source)
		})
	}
}

func TestWrap(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		msg  string
		opts []jettison.Option

		expectNil          bool
		expectedHopsCount  int
		expectedErrorCount int // in the latest hop
	}{
		{
			name:      "nil err",
			err:       nil,
			expectNil: true,
		},
		{
			name:               "non-Jettison err",
			err:                stdlib_errors.New("errors: first"),
			msg:                "errors: second",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
		},
		{
			name:               "Jettison err",
			err:                errors.New("errors: first"),
			msg:                "errors: second",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			line := nextLine()
			err := errors.Wrap(tc.err, tc.msg, tc.opts...)
			if tc.expectNil {
				assert.NoError(t, err)
				return
			}

			// We expect the returned error to be a Jettison error value.
			je, ok := err.(*errors.JettisonError)
			require.True(t, ok)

			assert.Len(t, je.Hops, tc.expectedHopsCount)
			assert.Len(t, je.Hops[0].Errors, tc.expectedErrorCount)
			assert.Equal(t, tc.msg, je.Hops[0].Errors[0].Message)

			assert.Equal(t,
				"github.com/luno/jettison/errors/errors_test.go:"+line,
				je.Hops[0].Errors[0].Source)
		})
	}
}

func TestIs(t *testing.T) {
	id1 := errors.New("id1", j.C("id1"))
	id2 := errors.New("id2", j.C("id2"))
	id3 := stdlib_errors.New("id3")

	testCases := []struct {
		name      string
		err       error
		target    error
		expResult bool
	}{
		{
			name:      "target nil returns false",
			err:       id1,
			expResult: false,
		},
		{
			name:      "err nil returns false",
			target:    id1,
			expResult: false,
		},
		{
			name:      "standard lib err return false",
			err:       stdlib_errors.New("err"),
			target:    errors.New("target"),
			expResult: false,
		},
		{
			name:      "standard lib target return false",
			err:       errors.New("err"),
			target:    stdlib_errors.New("target"),
			expResult: false,
		},
		{
			name:      "unrelated errors returns false",
			err:       errors.Wrap(id1, "random"),
			target:    id2,
			expResult: false,
		},
		{
			name:      "related errors returns true",
			err:       errors.Wrap(id1, "outer", j.C("outer")),
			target:    id1,
			expResult: true,
		},
		{
			name:      "target with no code returns false",
			err:       errors.New("err"),
			target:    errors.New("target"),
			expResult: false,
		},
		{
			name:      "equal error values return true",
			err:       id3,
			target:    id3,
			expResult: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			// Test Jettison's implementation of Is().
			assert.Equal(t, tc.expResult, errors.Is(tc.err, tc.target))

			// Test Go 2 spec's implementation of Is().
			assert.Equal(t, tc.expResult, xerrors.Is(tc.err, tc.target))
		})
	}
}

func TestIsAny(t *testing.T) {
	t1 := errors.New("t1", j.C("1"))
	t2 := errors.New("t2", j.C("2"))
	t3 := stdlib_errors.New("t3")
	e := errors.New("e", j.C("1"))

	assert.True(t, errors.IsAny(e, t1, t2, t3))
	assert.True(t, errors.IsAny(e, t1, t2))
	assert.False(t, errors.IsAny(e, t2, t3))
}

func TestGetCodes(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expCodes []string
	}{
		{
			name: "stdlib error returns nothing",
			err:  stdlib_errors.New("test"),
		},
		{
			name:     "unwrapped error returns its code",
			err:      errors.New("test", j.C("code")),
			expCodes: []string{"code"},
		},
		{
			name: "wrapped error returns both codes",
			err: errors.Wrap(errors.New("inner", j.C("inner")),
				"outer", j.C("outer")),
			expCodes: []string{"outer", "inner"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			codes := errors.GetCodes(tc.err)
			assert.Equal(t, tc.expCodes, codes)
		})
	}
}

// nextLine returns the next line after the caller.
func nextLine() string {
	_, _, line, _ := runtime.Caller(1)
	return strconv.Itoa(line + 1)
}

var errTest = errors.New("test error", j.C("ERR_59bed5816cb39f35"))

func TestIsUnwrap(t *testing.T) {
	err := errTest
	for i := 0; i < 5; i++ {
		err = errors.Wrap(err, "wrap", j.KV("i", i))
	}

	orig := err.Error()
	ok := errors.Is(err, errTest)
	require.True(t, ok)

	require.Equal(t, orig, err.Error())
}

// TestIsCompatibility tests that jettison and golang.org/exp/errors have Is()
// implementations with the same behaviour.
func TestIsCompatibility(t *testing.T) {
	matrix := make(map[int]map[int][3]bool)
	for i1 := 0; i1 < 3; i1++ {
		matrix[i1] = make(map[int][3]bool)
		for i2 := 0; i2 < 3; i2++ {
			matrix[i1][i2] = [3]bool{false, false, false}
		}
	}

	// index 0 - Jettison with an error code
	err1 := errors.New("err1", j.C("ERR_1"))
	err2 := errors.Wrap(err1, "err2")
	el := []error{nil, err1, err2}
	for i1, e1 := range el {
		for i2, e2 := range el {
			row := matrix[i1][i2]
			row[0] = errors.Is(e1, e2)
			matrix[i1][i2] = row
		}
	}

	// index 1 - Jettison without an error code
	err1 = errors.New("err1")
	err2 = errors.Wrap(err1, "err2")
	el = []error{nil, err1, err2}
	for i1, e1 := range el {
		for i2, e2 := range el {
			row := matrix[i1][i2]
			row[1] = errors.Is(e1, e2)
			matrix[i1][i2] = row
		}
	}

	// index 3 - golang.org/x/exp/errors
	err1 = xerrors.New("err1")
	err2 = xerrors.Errorf("err2: %w", err1)
	el = []error{nil, err1, err2}
	for i1, e1 := range el {
		for i2, e2 := range el {
			row := matrix[i1][i2]
			row[2] = xerrors.Is(e1, e2)
			matrix[i1][i2] = row
		}
	}

	// Each row in the compability matrix should have identical entries.
	for i1, submatrix := range matrix {
		for i2, row := range submatrix {
			assert.Equal(t, row[0], row[1], fmt.Sprintf("matrix[%d][%d] - jettison_code == jettison_no_code", i1, i2))
			assert.Equal(t, row[1], row[2], fmt.Sprintf("matrix[%d][%d] - jettison_no_code == xerrors", i1, i2))
		}
	}
}

// TestUnwrapCompatibility tests that jettison and golang.org/exp/errors have
// compatible wrapping/unwrapping of errors.
func TestUnwrapCompatibility(t *testing.T) {
	err1 := errors.New("err1")
	err2 := xerrors.Errorf("err2: %w", err1)
	err3 := errors.Wrap(err2, "err3", j.C("ERR_3"))
	err4 := xerrors.Errorf("err4: %w", err3)
	err5 := errors.Wrap(err4, "err5")

	// For testing code equality as well as value equality.
	err3Clone := errors.New("err3_clone", j.C("ERR_3"))

	assert.True(t, errors.Is(err5, err1))
	assert.True(t, errors.Is(err5, err2))
	assert.True(t, errors.Is(err5, err3))
	assert.True(t, errors.Is(err5, err3Clone))
	assert.True(t, errors.Is(err5, err4))
}

var ErrFoo = errors.New("foo", errors.WithoutStackTrace())

func TestWithoutStackTrace(t *testing.T) {
	je := ErrFoo.(*errors.JettisonError)
	require.Empty(t, je.Hops[0].StackTrace)

	// ErrFoo doesn't have stacktrace, but is has a source.
	source := je.Hops[0].Errors[0].Source
	require.True(t, strings.HasPrefix(source, "github.com/luno/jettison/errors/errors_test.go"))

	err := errors.Wrap(ErrFoo, "wrap adds stack trace")
	je = err.(*errors.JettisonError)
	require.Len(t, je.Hops, 1)
	require.Len(t, je.Hops[0].Errors, 2)
	require.NotEmpty(t, je.Hops[0].StackTrace)
}
