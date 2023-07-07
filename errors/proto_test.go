package errors_test

import (
	"io"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/internal/jettisonpb"
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
				&jettisonpb.WrappedError{Metadata: &jettisonpb.Metadata{Code: "abc"}},
			},
			expError: errors.JettisonError{Metadata: models.Metadata{Code: "abc"}},
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

			je, err := errors.FromStatus(s)
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
				Metadata: &jettisonpb.Metadata{
					Trace:     &jettisonpb.Hop{},
					KeyValues: []*jettisonpb.KeyValue{{Key: "key", Value: "value"}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := errors.ErrorToProto(tc.err)
			jtest.RequireNil(t, err)
			assert.Equal(t, tc.expProto, p)
		})
	}
}
