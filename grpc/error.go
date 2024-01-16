package grpc

import (
	"context"
	stderrors "errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/grpc/internal"
	"github.com/luno/jettison/grpc/internal/jettisonpb"
	"github.com/luno/jettison/models"
)

type Error struct {
	err error
	s   *status.Status
}

func (g Error) Error() string {
	return g.err.Error()
}

func (g Error) GRPCStatus() *status.Status {
	return g.s
}

func (g Error) Unwrap() error {
	return g.err
}

// Wrap will construct an Error that will serialise err when
// needed by gRPC by exposing the GRPCStatus method
// TODO(adam): Make generic with error instead of JettisonError
func Wrap(err *errors.JettisonError) Error {
	return Error{s: toStatus(err), err: err}
}

// FromStatus will deserialise the details from the status
// into an Error
func FromStatus(s *status.Status) Error {
	je, ok := fromStatus(s)
	if !ok {
		// NoReturnErr: If there's no error in the status, just return an empty error
		return Error{s: s}
	}
	return Error{s: s, err: je}
}

// toStatus marshals the given jettison error into a *grpc.Status object,
// with a message given by the most recently wrapped error in the list of
// hops.
func toStatus(je *errors.JettisonError) *status.Status {
	c := codes.Unknown
	if errors.Is(je.OriginalErr, context.Canceled) {
		c = codes.Canceled
	} else if errors.Is(je.OriginalErr, context.DeadlineExceeded) {
		c = codes.DeadlineExceeded
	}
	res := status.New(c, je.Message)

	for _, h := range je.Hops {
		hpb, err := internal.HopToProto(&h)
		if err != nil {
			log.Printf("jettison/errors: Failed to marshal hop to protobuf: %v", err)
			continue
		}

		withDetails, err := res.WithDetails(hpb)
		if err != nil {
			log.Printf("jettison/errors: Failed to add details to gRPC status: %v", err)
			continue
		}
		res = withDetails
	}

	withWrap, err := res.WithDetails(errorToProto(je))
	if err != nil {
		log.Printf("jettison/errors: Failed to add WrappedError to status: %v", err)
	} else {
		res = withWrap
	}
	return res
}

// fromStatus will unmarshal a *grpc.Status into a jettison error object,
// returning a nil error if and only if no unexpected details were found on the
// status.
func fromStatus(s *status.Status) (*errors.JettisonError, bool) {
	if s == nil {
		return nil, false
	}
	for _, d := range s.Details() {
		det, ok := d.(*jettisonpb.WrappedError)
		if !ok {
			continue
		}
		return errorFromProto(det), true
	}
	return nil, false
}

func errorFromProto(we *jettisonpb.WrappedError) *errors.JettisonError {
	je := &errors.JettisonError{
		Message:    we.Message,
		Binary:     we.Binary,
		Code:       we.Code,
		Source:     we.Source,
		StackTrace: we.StackTrace,
		KV:         kvFromProto(we.KeyValues),
	}
	if len(we.JoinedErrors) > 0 {
		var errs []error
		for _, joinErr := range we.JoinedErrors {
			errs = append(errs, errorFromProto(joinErr))
		}
		je.Err = stderrors.Join(errs...)
	} else if we.WrappedError != nil {
		je.Err = errorFromProto(we.WrappedError)
	}
	return je
}

func errorToProto(err error) *jettisonpb.WrappedError {
	if err == nil {
		return nil
	}
	var we jettisonpb.WrappedError
	je, ok := err.(*errors.JettisonError)
	if ok {
		we.Message = je.Message
		we.Binary = je.Binary
		we.Code = je.Code
		we.Source = je.Source
		we.StackTrace = je.StackTrace
		we.KeyValues = kvToProto(je.KV)
	}
	switch unw := err.(type) {
	case interface{ Unwrap() error }:
		we.WrappedError = errorToProto(unw.Unwrap())
	case interface{ Unwrap() []error }:
		for _, e := range unw.Unwrap() {
			we.JoinedErrors = append(we.JoinedErrors, errorToProto(e))
		}
	default:
		we.Message = err.Error()
	}
	return &we
}

func kvToProto(kvs []models.KeyValue) []*jettisonpb.KeyValue {
	if len(kvs) == 0 {
		return nil
	}
	res := make([]*jettisonpb.KeyValue, 0, len(kvs))
	for _, kv := range kvs {
		res = append(res, &jettisonpb.KeyValue{Key: kv.Key, Value: kv.Value})
	}
	return res
}

func kvFromProto(kvs []*jettisonpb.KeyValue) []models.KeyValue {
	if len(kvs) == 0 {
		return nil
	}
	res := make([]models.KeyValue, 0, len(kvs))
	for _, kv := range kvs {
		res = append(res, models.KeyValue{Key: kv.Key, Value: kv.Value})
	}
	return res
}
