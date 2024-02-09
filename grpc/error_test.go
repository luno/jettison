package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/grpc/internal/jettisonpb"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/jtest"
	"github.com/luno/jettison/models"
)

type source string

func (s source) ApplyToError(je *errors.JettisonError) {
	je.Source = string(s)
}

func TestFromStatus(t *testing.T) {
	testCases := []struct {
		name     string
		details  []proto.Message
		expJetty errors.JettisonError
		expOk    bool
	}{
		{
			name: "only a wrapped error",
			details: []proto.Message{
				&jettisonpb.WrappedError{Message: "test"},
			},
			expJetty: errors.JettisonError{Message: "test"},
			expOk:    true,
		},
		{
			name: "wrapped meta",
			details: []proto.Message{
				&jettisonpb.WrappedError{Code: "abc"},
			},
			expJetty: errors.JettisonError{Code: "abc"},
			expOk:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := status.New(codes.Unknown, "")
			s, err := s.WithDetails(tc.details...)
			jtest.RequireNil(t, err)

			je, ok := fromStatus(s)
			require.Equal(t, tc.expOk, ok)
			if ok {
				assert.Equal(t, tc.expJetty, *je)
			}
		})
	}
}

func TestToProto(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expProto *jettisonpb.WrappedError
	}{
		{
			name:     "std error",
			err:      io.EOF,
			expProto: &jettisonpb.WrappedError{Message: "EOF"},
		},
		{
			name:     "jettison error",
			err:      &errors.JettisonError{Message: "hi"},
			expProto: &jettisonpb.WrappedError{Message: "hi"},
		},
		{
			name: "wrapped error",
			err:  errors.Wrap(io.EOF, "hello", errors.WithoutStackTrace(), source(""), j.KV("key", "value")),
			expProto: &jettisonpb.WrappedError{
				Message: "hello",
				WrappedError: &jettisonpb.WrappedError{
					Message: "EOF",
				},
				KeyValues: []*jettisonpb.KeyValue{{Key: "key", Value: "value"}},
			},
		},
		{
			name:     "std error with non-utf 8",
			err:      fmt.Errorf("\xc5"),
			expProto: &jettisonpb.WrappedError{Message: "[snip]"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := errorToProto(tc.err)
			assert.Equal(t, tc.expProto, p)
		})
	}
}

func TestToFromStatus(t *testing.T) {
	errors.SetTraceConfigTesting(t, errors.TestingConfig)

	getStrconvErr := func() error {
		_, err := strconv.Atoi("nan")
		return err
	}

	testCases := []struct {
		name     string
		err      error
		expJetty errors.JettisonError
	}{
		{
			name: "single error, single param",
			err: errors.New("msg",
				j.KV("key", "value"),
				errors.WithoutStackTrace(),
			),
			expJetty: errors.JettisonError{
				Message: "msg",
				Source:  "error_test.go TestToFromStatus",
				KV: []models.KeyValue{
					{Key: "key", Value: "value"},
				},
			},
		},
		{
			name: "single error, many param",
			err: errors.New("msg", errors.WithoutStackTrace(),
				j.MKV{"key1": "value1", "key2": "value2"},
			),
			expJetty: errors.JettisonError{
				Message: "msg",
				Source:  "error_test.go TestToFromStatus",
				KV: []models.KeyValue{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
			},
		},
		{
			name: "wrapped error",
			err: errors.Wrap(
				errors.New("inner", errors.WithoutStackTrace()),
				"outer", errors.WithoutStackTrace(),
			),
			expJetty: errors.JettisonError{
				Message: "outer",
				Source:  "error_test.go TestToFromStatus",
				Err: &errors.JettisonError{
					Message: "inner",
					Source:  "error_test.go TestToFromStatus",
				},
			},
		},
		{
			name: "wrapped but not jetty",
			err: errors.Wrap(
				io.ErrUnexpectedEOF,
				"jetty", errors.WithoutStackTrace(),
			),
			expJetty: errors.JettisonError{
				Message: "jetty",
				Source:  "error_test.go TestToFromStatus",
				Err: &errors.JettisonError{
					Message: "unexpected EOF",
				},
			},
		},
		{
			name: "not jetty",
			err:  io.ErrUnexpectedEOF,
			expJetty: errors.JettisonError{
				Message: "unexpected EOF",
			},
		},
		{
			name: "all jetty details, recursive",
			err: &errors.JettisonError{
				Message:    "msg",
				Binary:     "binary",
				StackTrace: []string{"stack", "trace"},
				Code:       "code number 1",
				Source:     "sourcefile:line",
				KV: []models.KeyValue{
					{Key: "k1", Value: "v1"},
				},
				Err: &errors.JettisonError{
					Message:    "inner msg",
					Binary:     "binary2",
					StackTrace: []string{"hello", "world"},
					Code:       "code number 1",
					Source:     "sourcefile:line",
					KV: []models.KeyValue{
						{Key: "k2", Value: "v2"},
					},
				},
			},
			expJetty: errors.JettisonError{
				Message:    "msg",
				Binary:     "binary",
				StackTrace: []string{"stack", "trace"},
				Code:       "code number 1",
				Source:     "sourcefile:line",
				KV: []models.KeyValue{
					{Key: "k1", Value: "v1"},
				},
				Err: &errors.JettisonError{
					Message:    "inner msg",
					Binary:     "binary2",
					StackTrace: []string{"hello", "world"},
					Code:       "code number 1",
					Source:     "sourcefile:line",
					KV: []models.KeyValue{
						{Key: "k2", Value: "v2"},
					},
				},
			},
		},
		{
			name: "non-utf8 in strings",
			err: &errors.JettisonError{
				Message:    "msg\xc5",
				Binary:     "b\xc5in",
				StackTrace: []string{"\xc5 one", "two \xc5"},
				Code:       "code \xc5 here",
				Source:     "source \xc5 here",
				KV: []models.KeyValue{
					{
						Key:   "key with \xc5",
						Value: "value with \xc5",
					},
				},
			},
			expJetty: errors.JettisonError{
				Message:    "msg[snip]",
				Binary:     "b[snip]in",
				StackTrace: []string{"[snip] one", "two [snip]"},
				Code:       "code [snip] here",
				Source:     "source [snip] here",
				KV: []models.KeyValue{
					{
						Key:   "key with [snip]",
						Value: "value with [snip]",
					},
				},
			},
		},
		{
			name: "non-jettison but can unwrap, results in some redundant messages",
			err:  errors.Wrap(getStrconvErr(), "wrapper", errors.WithoutStackTrace()),
			expJetty: errors.JettisonError{
				Message: "wrapper",
				Source:  "error_test.go TestToFromStatus",
				Err: &errors.JettisonError{
					Message: "strconv.Atoi: parsing \"nan\": invalid syntax",
					Err: &errors.JettisonError{
						Message: "invalid syntax",
					},
				},
			},
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expJetty: errors.JettisonError{Message: context.DeadlineExceeded.Error()},
		},
		{
			name: "wrapped context deadline exceeded",
			err:  errors.Wrap(context.DeadlineExceeded, "", errors.WithoutStackTrace()),
			expJetty: errors.JettisonError{
				Source: "error_test.go TestToFromStatus",
				Err: &errors.JettisonError{
					Message: context.DeadlineExceeded.Error(),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			st := toStatus(tc.err)
			je, ok := fromStatus(st)
			assert.True(t, ok)
			errorEqual(t, &tc.expJetty, je)
		})
	}
}

func errorEqual(t *testing.T, exp, act *errors.JettisonError) {
	assert.Equal(t, exp.Message, act.Message)
	assert.Equal(t, exp.Binary, act.Binary)
	assert.Equal(t, exp.StackTrace, act.StackTrace)
	assert.Equal(t, exp.Code, act.Code)
	assert.Equal(t, exp.Source, act.Source)
	assert.Equal(t, exp.KV, act.KV)
	nextJe, ok := exp.Err.(*errors.JettisonError)
	if ok {
		errorEqual(t, nextJe, act.Err.(*errors.JettisonError))
	} else {
		jtest.Assert(t, exp.Err, act.Err)
	}
}

func TestGRPCStatus(t *testing.T) {
	testCases := []struct {
		name      string
		err       error
		expStatus *status.Status
	}{
		{
			name:      "new error",
			err:       errors.New("hello"),
			expStatus: status.New(codes.Unknown, "hello"),
		},
		{
			name:      "wrapped deadline exceeded error",
			err:       errors.Wrap(context.DeadlineExceeded, ""),
			expStatus: status.New(codes.DeadlineExceeded, ""),
		},
		{
			name:      "wrapped canceled error",
			err:       errors.Wrap(context.Canceled, ""),
			expStatus: status.New(codes.Canceled, ""),
		},
		{
			name:      "double wrapped",
			err:       errors.Wrap(errors.Wrap(context.Canceled, ""), ""),
			expStatus: status.New(codes.Canceled, ""),
		},
		{
			name:      "status error",
			err:       status.Error(codes.Unavailable, "oh no!"),
			expStatus: status.New(codes.Unavailable, "oh no!"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := outgoingError(tc.err)
			stater, ok := e.(interface{ GRPCStatus() *status.Status })
			require.True(t, ok)

			s := stater.GRPCStatus()

			assert.Equal(t, tc.expStatus.Code(), s.Code())
			assert.Equal(t, tc.expStatus.Message(), s.Message())
		})
	}
}

func TestErrorIs(t *testing.T) {
	testCases := []struct {
		name  string
		err   Error
		isErr error
		expIs bool
	}{
		{
			name: "deadline is",
			err: Error{
				s: status.New(codes.DeadlineExceeded, ""),
			},
			isErr: context.DeadlineExceeded,
			expIs: true,
		},
		{
			name: "canceled is",
			err: Error{
				s: status.New(codes.Canceled, ""),
			},
			isErr: context.Canceled,
			expIs: true,
		},
		{
			name: "deadline isn't canceled",
			err: Error{
				s: status.New(codes.DeadlineExceeded, ""),
			},
			isErr: context.Canceled,
			expIs: false,
		},
		{
			name: "unknown isn't canceled",
			err: Error{
				s: status.New(codes.Unknown, ""),
			},
			isErr: context.Canceled,
			expIs: false,
		},
		{
			name: "embedded canceled is unwrapped",
			err: Error{
				s:   status.New(codes.Unknown, ""),
				err: context.Canceled,
			},
			isErr: context.Canceled,
			expIs: true,
		},
		{
			name: "works with jettison codes",
			err: Error{
				err: errors.New("", j.C("codey mcCode face")),
			},
			isErr: errors.New("", j.C("codey mcCode face")),
			expIs: true,
		},
		{
			name: "works with jettison codes, negative",
			err: Error{
				err: errors.New("", j.C("not cool")),
			},
			isErr: errors.New("", j.C("codey mcCode face")),
			expIs: false,
		},
		{
			name:  "handles nil status",
			err:   Error{},
			isErr: context.Canceled,
			expIs: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			is := errors.Is(tc.err, tc.isErr)
			assert.Equal(t, tc.expIs, is)
		})
	}
}

func TestFromError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expError string
	}{
		{
			name:     "fmt error",
			err:      fmt.Errorf("hello, world"),
			expError: "hello, world",
		},
		{
			name:     "wrapped fmt",
			err:      fmt.Errorf("first: %v", fmt.Errorf("second")),
			expError: "first: second",
		},
		{
			name:     "jet wrapped fmt",
			err:      errors.Wrap(fmt.Errorf("inner"), "outer"),
			expError: "outer: inner",
		},
		{
			name:     "context error",
			err:      context.Canceled,
			expError: "context canceled",
		},
		{
			name:     "wrapped context",
			err:      errors.Wrap(context.Canceled, ""),
			expError: "context canceled",
		},
		{
			name:     "jetty",
			err:      errors.New("hello, jetty"),
			expError: "hello, jetty",
		},
		{
			name:     "wrapped jetty",
			err:      errors.Wrap(errors.New("hi"), "jet"),
			expError: "jet: hi",
		},
		{
			name:     "grpc error",
			err:      status.Error(codes.Unavailable, "message"),
			expError: "message",
		},
		{
			name:     "jetty, over the wire",
			err:      status.Convert(Wrap(errors.Wrap(errors.New("hey"), "yo"))).Err(),
			expError: "yo: hey",
		},
		{
			name:     "context, over the wire",
			err:      status.Convert(Wrap(context.Canceled)).Err(),
			expError: "context canceled",
		},
		{
			name:     "deadline, over the wire",
			err:      status.Convert(Wrap(context.DeadlineExceeded)).Err(),
			expError: "context deadline exceeded",
		},
		{
			name:     "sql, over the wire",
			err:      status.Convert(Wrap(sql.ErrNoRows)).Err(),
			expError: "sql: no rows in result set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := FromError(tc.err)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expError, err.Error())
		})
	}
}
