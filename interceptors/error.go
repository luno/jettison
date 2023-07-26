package interceptors

import (
	"context"
	stderrors "errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/internal/jettisonpb"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/models"
)

var ErrInvalidError = errors.New("jettison/errors: given grpc.Status does not contain a valid jettison error", j.C("ERR_e60c52eceb509f04"))

type gRPCError struct {
	err error
	s   *status.Status
}

func (g gRPCError) Error() string {
	return g.err.Error()
}

func (g gRPCError) GRPCStatus() *status.Status {
	return g.s
}

func (g gRPCError) Unwrap() error {
	return g.err
}

func gRPCWrap(je *errors.JettisonError) gRPCError {
	return gRPCError{s: toStatus(je), err: je}
}

// toStatus marshals the given jettison error into a *grpc.Status object,
// with a message given by the most recently wrapped error in the list of
// hops.
func toStatus(je *errors.JettisonError) *status.Status {
	msg := ""
	if le, ok := je.LatestError(); ok {
		msg = le.Message
	}

	c := codes.Unknown
	if errors.Is(je.OriginalErr, context.Canceled) {
		c = codes.Canceled
	} else if errors.Is(je.OriginalErr, context.DeadlineExceeded) {
		c = codes.DeadlineExceeded
	}

	res := status.New(c, msg)

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
	// TODO(adam): Include WrappedError details when services using FromStatus can handle it
	return res
}

// FromStatus unmarshals a *grpc.Status into a jettison error object,
// returning a nil error if and only if no unexpected details were found on the
// status.
func FromStatus(s *status.Status) (*errors.JettisonError, error) {
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
			je, err := ErrorFromProto(det)
			if err == nil {
				// TODO(adam): Just return je when we no longer rely on Hops
				res.Message = je.Message
				res.Binary = je.Binary
				res.Code = je.Code
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

func ErrorFromProto(we *jettisonpb.WrappedError) (*errors.JettisonError, error) {
	je := &errors.JettisonError{
		Message:    we.Message,
		Binary:     we.Binary,
		Code:       we.Code,
		StackTrace: we.StackTrace,
		KV:         kvFromProto(we.KeyValues),
	}
	if len(we.JoinedErrors) > 0 {
		var errs []error
		for _, joinErr := range we.JoinedErrors {
			subErr, err := ErrorFromProto(joinErr)
			if err != nil {
				return nil, err
			}
			errs = append(errs, subErr)
		}
		je.Err = stderrors.Join(errs...)
	} else if we.WrappedError != nil {
		subErr, err := ErrorFromProto(we.WrappedError)
		if err != nil {
			return nil, err
		}
		je.Err = subErr
	}
	return je, nil
}

func ErrorToProto(err error) (*jettisonpb.WrappedError, error) {
	if err == nil {
		return nil, nil
	}
	var we jettisonpb.WrappedError
	je, ok := err.(*errors.JettisonError)
	if ok {
		we.Message = je.Message
		we.Binary = je.Binary
		we.Code = je.Code
		we.StackTrace = je.StackTrace
		we.KeyValues = kvToProto(je.KV)
	}
	switch unw := err.(type) {
	case interface{ Unwrap() error }:
		subWe, err := ErrorToProto(unw.Unwrap())
		if err != nil {
			return nil, err
		}
		we.WrappedError = subWe
	case interface{ Unwrap() []error }:
		for _, e := range unw.Unwrap() {
			subWe, err := ErrorToProto(e)
			if err != nil {
				return nil, err
			}
			we.JoinedErrors = append(we.JoinedErrors, subWe)
		}
	default:
		we.Message = err.Error()
	}
	return &we, nil
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
