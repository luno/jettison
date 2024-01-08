package errors_test

import (
	stdlib_errors "errors"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			err := errors.New(tc.msg, tc.opts...).(*errors.JettisonError)

			assert.Equal(t, tc.msg, err.Message)
			assert.Equal(t,
				"github.com/luno/jettison/errors/errors_test.go:"+line,
				err.Source,
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

		expectNil       bool
		expectedMessage string
	}{
		{
			name:      "nil err",
			err:       nil,
			expectNil: true,
		},
		{
			name:            "non-Jettison err",
			err:             stdlib_errors.New("errors: first"),
			msg:             "errors: second",
			expectedMessage: "errors: second: errors: first",
		},
		{
			name:            "Jettison err",
			err:             errors.New("errors: first"),
			msg:             "errors: second",
			expectedMessage: "errors: second: errors: first",
		},
		{
			name:            "wrap empty message",
			err:             errors.New("test value"),
			msg:             "",
			expectedMessage: "test value",
		},
		{
			name:            "wrap empty message, with stdlib error",
			err:             stdlib_errors.New("test value"),
			msg:             "",
			expectedMessage: "test value",
		},
		{
			name:            "wrap known error",
			err:             io.EOF,
			msg:             "end of file",
			expectedMessage: "end of file: EOF",
		},
		{
			name:            "wrap options message, ignores options",
			err:             errors.New("test value", j.KV("key", "value")),
			msg:             "hello",
			expectedMessage: "hello: test value",
		},
		{
			name:            "wrap wrapped message",
			err:             errors.Wrap(errors.New("test value"), "world"),
			msg:             "hello",
			expectedMessage: "hello: world: test value",
		},
		{
			name:            "double empty wrapped message",
			err:             errors.Wrap(errors.New("test value"), ""),
			msg:             "",
			expectedMessage: "test value",
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
			je := err.(*errors.JettisonError)
			assert.Equal(t, tc.msg, je.Message)
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
	errNoCode := errors.New("err_no_code")

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
			name:      "err is self",
			err:       id1,
			target:    id1,
			expResult: true,
		},
		{
			name:      "std err is self",
			err:       id3,
			target:    id3,
			expResult: true,
		},
		{
			name:      "no code is self",
			err:       errNoCode,
			target:    errNoCode,
			expResult: true,
		},
		{
			name:      "new same message is equal",
			err:       errors.New("hello, world"),
			target:    errors.New("hello, world"),
			expResult: true,
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
		err = errors.Wrap(err, "wrap").(*errors.JettisonError)
	}

	orig := err.Error()
	ok := errors.Is(err, errTest)
	require.True(t, ok)

	require.Equal(t, orig, err.Error())
}

func TestWithoutStackTrace(t *testing.T) {
	errFoo := errors.New("foo", errors.WithoutStackTrace()).(*errors.JettisonError)
	assert.Empty(t, errFoo.StackTrace)
	assert.Empty(t, errFoo.Source)

	err := errors.Wrap(errFoo, "wrap adds stack trace").(*errors.JettisonError)
	assert.NotEmpty(t, err.StackTrace)
	assert.NotEmpty(t, err.Source)
}

func TestErrorMetadata(t *testing.T) {
	testCases := []struct {
		name       string
		err        *errors.JettisonError
		expError   errors.JettisonError
		expNoTrace bool
	}{
		{
			name: "new kv",
			err:  errors.New("one", j.KV("test", "val")).(*errors.JettisonError),
			expError: errors.JettisonError{
				Message: "one",
				KV:      []models.KeyValue{{Key: "test", Value: "val"}},
			},
		},
		{
			name: "new code",
			err:  errors.New("one", errors.WithCode("code")).(*errors.JettisonError),
			expError: errors.JettisonError{
				Message: "one", Code: "code",
			},
		},
		{
			name:       "without stacktrace",
			err:        errors.New("one", errors.WithoutStackTrace()).(*errors.JettisonError),
			expNoTrace: true,
			expError:   errors.JettisonError{Message: "one"},
		},
		{
			name:     "wrap non-jettison, gets a trace",
			err:      errors.Wrap(io.EOF, "hi").(*errors.JettisonError),
			expError: errors.JettisonError{Message: "hi"},
		},
		{
			name: "wrap non-jettison, with kv",
			err:  errors.Wrap(io.EOF, "hi", j.KV("key", "value")).(*errors.JettisonError),
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
			).(*errors.JettisonError),
			expError: errors.JettisonError{
				Message: "outer",
				KV:      []models.KeyValue{{Key: "outer", Value: "outer_value"}},
			},
			expNoTrace: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expNoTrace {
				assert.Empty(t, tc.err.Binary)
				assert.Empty(t, tc.err.StackTrace)
			} else {
				assert.NotEmpty(t, tc.err.Binary)
				assert.NotEmpty(t, tc.err.StackTrace)
			}

			assert.Equal(t, tc.expError.Message, tc.err.Message)
			assert.Equal(t, tc.expError.Code, tc.err.Code)
			assert.Equal(t, tc.expError.KV, tc.err.KV)
		})
	}
}

func TestWithStacktrace(t *testing.T) {
	base := errors.New("base").(*errors.JettisonError)
	assert.NotEmpty(t, base.StackTrace)

	// No stack trace if base error has one already
	wrapped := errors.Wrap(base, "wrap").(*errors.JettisonError)
	assert.Empty(t, wrapped.StackTrace)

	// Get trace if explicitly requested
	wst := errors.Wrap(base, "stacky", errors.WithStackTrace()).(*errors.JettisonError)
	assert.NotEmpty(t, wst.StackTrace)
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
