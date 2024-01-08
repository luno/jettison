package grpc

import (
	"context"
	"io"
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

func TestFromStatus(t *testing.T) {
	testCases := []struct {
		name     string
		details  []proto.Message
		expError errors.JettisonError
	}{
		{
			name: "one hop this time",
			details: []proto.Message{
				&jettisonpb.Hop{Binary: "mc hammer"},
			},
			expError: errors.JettisonError{
				Hops: []models.Hop{{Binary: "mc hammer"}},
			},
		},
		{
			name: "only a wrapped error",
			details: []proto.Message{
				&jettisonpb.WrappedError{Message: "test"},
			},
			expError: errors.JettisonError{Message: "test"},
		},
		{
			name: "wrapped meta",
			details: []proto.Message{
				&jettisonpb.WrappedError{Code: "abc"},
			},
			expError: errors.JettisonError{Code: "abc"},
		},
		{
			name: "still decode Hop when WrappedError is there",
			details: []proto.Message{
				&jettisonpb.Hop{Binary: "mc hammer"},
				&jettisonpb.WrappedError{},
			},
			expError: errors.JettisonError{
				Hops: []models.Hop{{Binary: "mc hammer"}},
			},
		},
		{
			name: "fully hopped",
			details: []proto.Message{
				&jettisonpb.Hop{
					Binary:     "binny",
					StackTrace: []string{"a", "b", "c"},
					Errors: []*jettisonpb.Error{
						{
							Code:    "error1",
							Message: "msg1",
							Source:  "anywhere",
							Parameters: []*jettisonpb.KeyValue{
								{Key: "test_key_1", Value: "test_value_1"},
							},
						},
						{
							Code:    "error2",
							Message: "msg2",
							Source:  "somewhere else",
							Parameters: []*jettisonpb.KeyValue{
								{Key: "test_key_2", Value: "test_value_2"},
							},
						},
					},
				},
			},
			expError: errors.JettisonError{Hops: []models.Hop{
				{
					Binary:     "binny",
					StackTrace: []string{"a", "b", "c"},
					Errors: []models.Error{
						{
							Code:    "error1",
							Message: "msg1",
							Source:  "anywhere",
							Parameters: []models.KeyValue{
								{Key: "test_key_1", Value: "test_value_1"},
							},
						},
						{
							Code:    "error2",
							Message: "msg2",
							Source:  "somewhere else",
							Parameters: []models.KeyValue{
								{Key: "test_key_2", Value: "test_value_2"},
							},
						},
					},
				},
			}},
		},
		{
			name: "multi hops",
			details: []proto.Message{
				&jettisonpb.Hop{Binary: "bin1"},
				&jettisonpb.Hop{Binary: "bin2"},
			},
			expError: errors.JettisonError{
				Hops: []models.Hop{{Binary: "bin1"}, {Binary: "bin2"}},
			},
		},
		{
			name: "multi hops with a wrapper",
			details: []proto.Message{
				&jettisonpb.Hop{Binary: "bin1"},
				&jettisonpb.WrappedError{Message: "hello"},
				&jettisonpb.Hop{Binary: "bin2"},
			},
			expError: errors.JettisonError{
				Hops:    []models.Hop{{Binary: "bin1"}, {Binary: "bin2"}},
				Message: "hello",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := status.New(codes.Unknown, "")
			s, err := s.WithDetails(tc.details...)
			jtest.RequireNil(t, err)

			je, err := fromStatus(s)
			jtest.RequireNil(t, err)

			assert.Equal(t, tc.expError, *je)
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
			err:  errors.Wrap(io.EOF, "hello", errors.WithoutStackTrace(), j.KV("key", "value")),
			expProto: &jettisonpb.WrappedError{
				Message: "hello",
				WrappedError: &jettisonpb.WrappedError{
					Message: "EOF",
				},
				KeyValues: []*jettisonpb.KeyValue{{Key: "key", Value: "value"}},
			},
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
	testCases := []struct {
		name   string
		errors []models.Error
	}{
		{
			name: "single error, single param",
			errors: []models.Error{
				{
					Message: "msg",
					Source:  "source",
					Parameters: []models.KeyValue{
						{Key: "key", Value: "value"},
					},
				},
			},
		},
		{
			name: "single error, many param",
			errors: []models.Error{
				{
					Message: "msg",
					Source:  "source",
					Parameters: []models.KeyValue{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
		},
		{
			name: "many error, many param",
			errors: []models.Error{
				{
					Message: "msg1",
					Source:  "source1",
					Parameters: []models.KeyValue{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
				{
					Message: "msg2",
					Source:  "source2",
					Parameters: []models.KeyValue{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			srvJe := &errors.JettisonError{Hops: []models.Hop{
				{Binary: "service", Errors: tc.errors},
			}}

			// Simulate going over the wire.
			st, ok := status.FromError(outgoingError(srvJe))
			require.True(t, ok)
			cliJe, err := fromStatus(st)
			jtest.RequireNil(t, err)

			assert.Equal(t, srvJe.Message, cliJe.Message)
			assert.Equal(t, srvJe.Binary, cliJe.Binary)
			assert.Equal(t, srvJe.StackTrace, cliJe.StackTrace)
			assert.Equal(t, srvJe.Code, cliJe.Code)
			assert.Equal(t, srvJe.Source, cliJe.Source)
			assert.Equal(t, srvJe.KV, cliJe.KV)
		})
	}
}

func TestNonUTF8CharsInHop(t *testing.T) {
	err := errors.JettisonError{
		Hops: []models.Hop{
			{
				Binary: "service",
				Errors: []models.Error{
					{
						Message: "msg1",
						Source:  "source1",
						Parameters: []models.KeyValue{
							{Key: "key1", Value: "a\xc5z"},
						},
					},
				},
			},
		},
	}
	assert.NotNil(t, outgoingError(&err))
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
