package errors

import (
	stdlib_errors "errors"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/luno/jettison"
	"github.com/luno/jettison/models"
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
			err := New(tc.msg, tc.opts...)

			je, ok := err.(*JettisonError)
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
		expectedMessage    string
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
			expectedMessage:    "errors: second: errors: first",
		},
		{
			name:               "Jettison err",
			err:                New("errors: first"),
			msg:                "errors: second",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
			expectedMessage:    "errors: second: errors: first",
		},
		{
			name:               "wrap empty message",
			err:                New("test value"),
			msg:                "",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
			expectedMessage:    "test value",
		},
		{
			name:               "wrap empty message, with stdlib error",
			err:                stdlib_errors.New("test value"),
			msg:                "",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
			expectedMessage:    "test value",
		},
		{
			name:               "wrap known error",
			err:                io.EOF,
			msg:                "end of file",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
			expectedMessage:    "end of file: EOF",
		},
		{
			name:               "wrap options message, ignores options",
			err:                New("test value", jettison.WithKeyValueString("key", "value")),
			msg:                "hello",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
			expectedMessage:    "hello: test value",
		},
		{
			name:               "wrap wrapped message",
			err:                Wrap(New("test value"), "world"),
			msg:                "hello",
			expectedHopsCount:  1,
			expectedErrorCount: 3,
			expectedMessage:    "hello: world: test value",
		},
		{
			name:               "double empty wrapped message",
			err:                Wrap(New("test value"), ""),
			msg:                "",
			expectedHopsCount:  1,
			expectedErrorCount: 3,
			expectedMessage:    "test value",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			line := nextLine()
			err := Wrap(tc.err, tc.msg, tc.opts...)
			if tc.expectNil {
				assert.NoError(t, err)
				return
			}

			// We expect the returned error to be a Jettison error value.
			je, ok := err.(*JettisonError)
			require.True(t, ok)

			assert.Len(t, je.Hops, tc.expectedHopsCount)
			assert.Len(t, je.Hops[0].Errors, tc.expectedErrorCount)
			assert.Equal(t, tc.msg, je.Hops[0].Errors[0].Message)

			assert.Equal(t,
				"github.com/luno/jettison/errors/errors_test.go:"+line,
				je.Hops[0].Errors[0].Source)

			assert.Equal(t, tc.expectedMessage, err.Error())
		})
	}
}

func TestIs(t *testing.T) {
	id1 := New("id1", WithCode("id1"))
	id2 := New("id2", WithCode("id2"))
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
			target:    New("target"),
			expResult: false,
		},
		{
			name:      "standard lib target return false",
			err:       New("err"),
			target:    stdlib_errors.New("target"),
			expResult: false,
		},
		{
			name:      "unrelated errors returns false",
			err:       Wrap(id1, "random"),
			target:    id2,
			expResult: false,
		},
		{
			name:      "related errors returns true",
			err:       Wrap(id1, "outer", WithCode("outer")),
			target:    id1,
			expResult: true,
		},
		{
			name:      "target with no code returns false",
			err:       New("err"),
			target:    New("target"),
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
			assert.Equal(t, tc.expResult, Is(tc.err, tc.target))

			// Test Go 2 spec's implementation of Is().
			assert.Equal(t, tc.expResult, xerrors.Is(tc.err, tc.target))
		})
	}
}

func TestIsAny(t *testing.T) {
	t1 := New("t1", WithCode("1"))
	t2 := New("t2", WithCode("2"))
	t3 := stdlib_errors.New("t3")
	e := New("e", WithCode("1"))

	assert.True(t, IsAny(e, t1, t2, t3))
	assert.True(t, IsAny(e, t1, t2))
	assert.False(t, IsAny(e, t2, t3))
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
			err:      New("test", WithCode("code")),
			expCodes: []string{"code"},
		},
		{
			name: "wrapped error returns both codes",
			err: Wrap(New("inner", WithCode("inner")),
				"outer", WithCode("outer")),
			expCodes: []string{"outer", "inner"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			codes := GetCodes(tc.err)
			assert.Equal(t, tc.expCodes, codes)
		})
	}
}

// nextLine returns the next line after the caller.
func nextLine() string {
	_, _, line, _ := runtime.Caller(1)
	return strconv.Itoa(line + 1)
}

var errTest = New("test error", WithCode("ERR_59bed5816cb39f35"))

func TestIsUnwrap(t *testing.T) {
	err := errTest
	for i := 0; i < 5; i++ {
		err = Wrap(err, "wrap")
	}

	orig := err.Error()
	ok := Is(err, errTest)
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
	err1 := New("err1", WithCode("ERR_1"))
	err2 := Wrap(err1, "err2")
	el := []error{nil, err1, err2}
	for i1, e1 := range el {
		for i2, e2 := range el {
			row := matrix[i1][i2]
			row[0] = Is(e1, e2)
			matrix[i1][i2] = row
		}
	}

	// index 1 - Jettison without an error code
	err1 = New("err1")
	err2 = Wrap(err1, "err2")
	el = []error{nil, err1, err2}
	for i1, e1 := range el {
		for i2, e2 := range el {
			row := matrix[i1][i2]
			row[1] = Is(e1, e2)
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
	err1 := New("err1")
	err2 := xerrors.Errorf("err2: %w", err1)
	err3 := Wrap(err2, "err3", WithCode("ERR_3"))
	err4 := xerrors.Errorf("err4: %w", err3)
	err5 := Wrap(err4, "err5")

	// For testing code equality as well as value equality.
	err3Clone := New("err3_clone", WithCode("ERR_3"))

	assert.True(t, Is(err5, err1))
	assert.True(t, Is(err5, err2))
	assert.True(t, Is(err5, err3))
	assert.True(t, Is(err5, err3Clone))
	assert.True(t, Is(err5, err4))
}

func TestWithoutStackTrace(t *testing.T) {
	errFoo := New("foo", WithoutStackTrace())

	je := errFoo.(*JettisonError)
	require.Empty(t, je.Hops[0].StackTrace)

	// ErrFoo doesn't have stacktrace, but is has a source.
	source := je.Hops[0].Errors[0].Source
	require.True(t, strings.HasPrefix(source, "github.com/luno/jettison/errors/errors_test.go"))

	err := Wrap(errFoo, "wrap adds stack trace")
	je = err.(*JettisonError)
	require.Len(t, je.Hops, 1)
	require.Len(t, je.Hops[0].Errors, 2)
	require.NotEmpty(t, je.Hops[0].StackTrace)
}

func TestErrorMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		expMetadata models.Metadata
		expNoTrace  bool
	}{
		{
			name: "new kv",
			err:  New("one", jettison.WithKeyValueString("test", "val")),
			expMetadata: models.Metadata{
				KV: []models.KeyValue{{Key: "test", Value: "val"}},
			},
		},
		{
			name: "new code",
			err:  New("one", WithCode("code")),
			expMetadata: models.Metadata{
				Code: "code",
			},
		},
		{
			name:       "without stacktrace",
			err:        New("one", WithoutStackTrace()),
			expNoTrace: true,
		},
		{
			name: "wrap non-jettison, gets a trace",
			err:  Wrap(io.EOF, "hi"),
		},
		{
			name: "wrap non-jettison, with kv",
			err:  Wrap(io.EOF, "hi", jettison.WithKeyValueString("key", "value")),
			expMetadata: models.Metadata{
				KV: []models.KeyValue{{Key: "key", Value: "value"}},
			},
		},
		{
			name: "wrapped with other options",
			err: Wrap(
				New("inner", jettison.WithKeyValueString("inner", "inner_value")),
				"outer",
				jettison.WithKeyValueString("outer", "outer_value"),
			),
			expMetadata: models.Metadata{
				KV: []models.KeyValue{{Key: "outer", Value: "outer_value"}},
			},
			expNoTrace: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			je := tc.err.(*JettisonError)
			if tc.expNoTrace {
				assert.Empty(t, je.Metadata.Trace.Binary)
				assert.Empty(t, je.Metadata.Trace.StackTrace)
			} else {
				assert.NotEmpty(t, je.Metadata.Trace.Binary)
				assert.NotEmpty(t, je.Metadata.Trace.StackTrace)
			}
			// Clear trace
			je.Metadata.Trace = models.Hop{}
			assert.Equal(t, tc.expMetadata, je.Metadata)
		})
	}
}
