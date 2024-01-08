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
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/models"
)

var ErrInvalidError = errors.New("jettison/errors: given grpc.Status does not contain a valid jettison error", j.C("ERR_e60c52eceb509f04"))

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
	je, err := fromStatus(s)
	if err != nil {
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
func fromStatus(s *status.Status) (*errors.JettisonError, error) {
	if s == nil {
		return nil, errors.Wrap(ErrInvalidError, "")
	} else if len(s.Details()) == 0 {
		return nil, errors.Wrap(ErrInvalidError, "")
	}

	var res errors.JettisonError
	for _, d := range s.Details() {
		switch det := d.(type) {
		case *jettisonpb.Hop:
			hop, err := internal.HopFromProto(det)
			if err != nil {
				return nil, err
			}
			res.Hops = append(res.Hops, *hop)
		case *jettisonpb.WrappedError:
			je, err := errorFromProto(det)
			if err == nil {
				// TODO(adam): Just return je when we no longer rely on Hops
				res.Message = je.Message
				res.Binary = je.Binary
				res.Code = je.Code
				res.Source = je.Source
				res.StackTrace = je.StackTrace
				res.KV = je.KV
				res.Err = je.Err
			}
		}
	}
	if len(res.Hops) == 0 && res.IsZero() {
		return nil, errors.Wrap(ErrInvalidError, "")
	}
	return &res, nil
}

func errorFromProto(we *jettisonpb.WrappedError) (*errors.JettisonError, error) {
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
			subErr, err := errorFromProto(joinErr)
			if err != nil {
				return nil, err
			}
			errs = append(errs, subErr)
		}
		je.Err = stderrors.Join(errs...)
	} else if we.WrappedError != nil {
		subErr, err := errorFromProto(we.WrappedError)
		if err != nil {
			return nil, err
		}
		je.Err = subErr
	}
	return je, nil
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
