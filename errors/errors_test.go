package errors_test

import (
	stdlib_errors "errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/models"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name string
		msg  string
		opts []errors.Option
	}{
		{
			name: "key/value setting",
			msg:  "key/value setting",
			opts: []errors.Option{
				j.KV("key", "value"),
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

			assert.Equal(t,
				"github.com/luno/jettison/errors/errors_test.go:"+line,
				je.Source,
			)
		})
	}
}

func TestWrap(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		msg  string
		opts []errors.Option

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
			err:                errors.New("errors: first"),
			msg:                "errors: second",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
			expectedMessage:    "errors: second: errors: first",
		},
		{
			name:               "wrap empty message",
			err:                errors.New("test value"),
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
			err:                errors.New("test value", j.KV("key", "value")),
			msg:                "hello",
			expectedHopsCount:  1,
			expectedErrorCount: 2,
			expectedMessage:    "hello: test value",
		},
		{
			name:               "wrap wrapped message",
			err:                errors.Wrap(errors.New("test value"), "world"),
			msg:                "hello",
			expectedHopsCount:  1,
			expectedErrorCount: 3,
			expectedMessage:    "hello: world: test value",
		},
		{
			name:               "double empty wrapped message",
			err:                errors.Wrap(errors.New("test value"), ""),
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
				je.Hops[0].Errors[0].Source,
			)
			assert.Equal(t,
				"github.com/luno/jettison/errors/errors_test.go:"+line,
				je.Source,
			)

			assert.Equal(t, tc.expectedMessage, err.Error())
		})
	}
}

func TestIs(t *testing.T) {
	id1 := errors.New("id1", errors.WithCode("id1"))
	id2 := errors.New("id2", errors.WithCode("id2"))
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
			err:       errors.Wrap(id1, "outer", errors.WithCode("outer")),
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
	t1 := errors.New("t1", errors.WithCode("1"))
	t2 := errors.New("t2", errors.WithCode("2"))
	t3 := stdlib_errors.New("t3")
	e := errors.New("e", errors.WithCode("1"))

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
			err:      errors.New("test", errors.WithCode("code")),
			expCodes: []string{"code"},
		},
		{
			name: "wrapped error returns both codes",
			err: errors.Wrap(errors.New("inner", errors.WithCode("inner")),
				"outer", errors.WithCode("outer")),
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

var errTest = errors.New("test error", errors.WithCode("ERR_59bed5816cb39f35"))

func TestIsUnwrap(t *testing.T) {
	err := errTest
	for i := 0; i < 5; i++ {
		err = errors.Wrap(err, "wrap")
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
	err1 := errors.New("err1", errors.WithCode("ERR_1"))
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
	err3 := errors.Wrap(err2, "err3", errors.WithCode("ERR_3"))
	err4 := xerrors.Errorf("err4: %w", err3)
	err5 := errors.Wrap(err4, "err5")

	// For testing code equality as well as value equality.
	err3Clone := errors.New("err3_clone", errors.WithCode("ERR_3"))

	assert.True(t, errors.Is(err5, err1))
	assert.True(t, errors.Is(err5, err2))
	assert.True(t, errors.Is(err5, err3))
	assert.True(t, errors.Is(err5, err3Clone))
	assert.True(t, errors.Is(err5, err4))
}

func TestWithoutStackTrace(t *testing.T) {
	errFoo := errors.New("foo", errors.WithoutStackTrace())

	je := errFoo.(*errors.JettisonError)
	require.Empty(t, je.Hops[0].StackTrace)

	// ErrFoo doesn't have stacktrace, but is has a source.
	source := je.Hops[0].Errors[0].Source
	require.True(t, strings.HasPrefix(source, "github.com/luno/jettison/errors/errors_test.go"))

	err := errors.Wrap(errFoo, "wrap adds stack trace")
	je = err.(*errors.JettisonError)
	require.Len(t, je.Hops, 1)
	require.Len(t, je.Hops[0].Errors, 2)
	require.NotEmpty(t, je.Hops[0].StackTrace)
}

func TestErrorMetadata(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		expError   errors.JettisonError
		expNoTrace bool
	}{
		{
			name: "new kv",
			err:  errors.New("one", j.KV("test", "val")),
			expError: errors.JettisonError{
				Message: "one",
				KV:      []models.KeyValue{{Key: "test", Value: "val"}},
			},
		},
		{
			name: "new code",
			err:  errors.New("one", errors.WithCode("code")),
			expError: errors.JettisonError{
				Message: "one", Code: "code",
			},
		},
		{
			name:       "without stacktrace",
			err:        errors.New("one", errors.WithoutStackTrace()),
			expNoTrace: true,
			expError:   errors.JettisonError{Message: "one"},
		},
		{
			name:     "wrap non-jettison, gets a trace",
			err:      errors.Wrap(io.EOF, "hi"),
			expError: errors.JettisonError{Message: "hi"},
		},
		{
			name: "wrap non-jettison, with kv",
			err:  errors.Wrap(io.EOF, "hi", j.KV("key", "value")),
			expError: errors.JettisonError{
				Message: "hi",
				KV:      []models.KeyValue{{Key: "key", Value: "value"}},
			},
		},
		{
			name: "wrapped with other options",
			err: errors.Wrap(
				errors.New("inner", j.KV("inner", "inner_value")),
				"outer",
				j.KV("outer", "outer_value"),
			),
			expError: errors.JettisonError{
				Message: "outer",
				KV:      []models.KeyValue{{Key: "outer", Value: "outer_value"}},
			},
			expNoTrace: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			je := tc.err.(*errors.JettisonError)
			if tc.expNoTrace {
				assert.Empty(t, je.Binary)
				assert.Empty(t, je.StackTrace)
			} else {
				assert.NotEmpty(t, je.Binary)
				assert.NotEmpty(t, je.StackTrace)
			}

			assert.Equal(t, tc.expError.Message, je.Message)
			assert.Equal(t, tc.expError.Code, je.Code)
			assert.Equal(t, tc.expError.KV, je.KV)
		})
	}
}

func TestWithStacktrace(t *testing.T) {
	base := errors.New("base")
	assert.NotEmpty(t, base.(*errors.JettisonError).StackTrace)

	// No stack trace if base error has one already
	wrapped := errors.Wrap(base, "wrap")
	assert.Empty(t, wrapped.(*errors.JettisonError).StackTrace)

	// Get trace if explicitly requested
	wst := errors.Wrap(base, "stacky", errors.WithStackTrace())
	assert.NotEmpty(t, wst.(*errors.JettisonError).StackTrace)
}

func TestWalk(t *testing.T) {
	testCases := []struct {
		name      string
		err       error
		stopN     int
		expErrors []string
	}{
		{name: "nil"},
		{
			name:      "simple error",
			err:       io.ErrUnexpectedEOF,
			expErrors: []string{io.ErrUnexpectedEOF.Error()},
		},
		{
			name: "wrapped",
			err:  errors.Wrap(errors.Wrap(errors.New("hello"), "inner"), "outer"),
			expErrors: []string{
				"outer: inner: hello",
				"inner: hello",
				"hello",
			},
		},
		{
			name:  "wrapped, stop early",
			err:   errors.Wrap(errors.Wrap(errors.New("hello"), "inner"), "outer"),
			stopN: 2,
			expErrors: []string{
				"outer: inner: hello",
				"inner: hello",
			},
		},
		{
			name: "joined",
			err: stdlib_errors.Join(
				errors.New("error one"),
				errors.New("error two"),
			),
			expErrors: []string{
				"error one\nerror two",
				"error one",
				"error two",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var errCount int
			var msgs []string
			errors.Walk(tc.err, func(err error) bool {
				msgs = append(msgs, err.Error())
				errCount++
				return tc.stopN == 0 || errCount < tc.stopN
			})
			assert.Equal(t, tc.expErrors, msgs)
		})
	}
}

func TestFlatten(t *testing.T) {
	err := errors.Wrap(
		stdlib_errors.Join(
			errors.Wrap(io.EOF, "wrapped"),
			errors.New("jet"),
			http.ErrNoCookie,
		),
		"outer",
	)
	act := errors.Flatten(err)
	exp := [][]string{
		{
			"outer: wrapped: EOF\njet\nhttp: named cookie not present",
			"wrapped: EOF\njet\nhttp: named cookie not present",
			"wrapped: EOF",
			"EOF",
		},
		{
			"outer: wrapped: EOF\njet\nhttp: named cookie not present",
			"wrapped: EOF\njet\nhttp: named cookie not present",
			"jet",
		},
		{
			"outer: wrapped: EOF\njet\nhttp: named cookie not present",
			"wrapped: EOF\njet\nhttp: named cookie not present",
			"http: named cookie not present",
		},
	}
	var msgs [][]string
	for _, p := range act {
		var pm []string
		for _, e := range p {
			pm = append(pm, e.Error())
		}
		msgs = append(msgs, pm)
	}
	assert.Equal(t, exp, msgs)
}
