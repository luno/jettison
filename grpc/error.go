package grpc

import (
	"context"
	stderrors "errors"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/grpc/internal/jettisonpb"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/models"
)

// Error wraps an error and a status.
// For outgoing errors, it allows us to create the grpc status and store it until needed.
// For incoming errors, we can store the original status from grpc, so we can check it later for certain codes
// this allows us to communicate context errors across grpc.
type Error struct {
	err error
	s   *status.Status
}

func (g Error) Error() string {
	return g.err.Error()
}

func (g Error) Is(target error) bool {
	if target == context.Canceled {
		return g.s.Code() == codes.Canceled
	} else if target == context.DeadlineExceeded {
		return g.s.Code() == codes.DeadlineExceeded
	}
	return false
}

func (g Error) GRPCStatus() *status.Status {
	return g.s
}

func (g Error) Unwrap() error {
	return g.err
}

// Wrap will construct an Error that will serialise err when
// needed by gRPC by exposing the GRPCStatus method
func Wrap(err error) Error {
	return Error{s: toStatus(err), err: err}
}

// FromStatus will de-serialise the details from the status
// into an Error
func FromStatus(s *status.Status) Error {
	je, ok := fromStatus(s)
	if !ok {
		e := errors.New("grpc status error",
			j.MKV{"code": s.Code(), "message": s.Message()},
			errors.WithoutStackTrace(),
		)
		return Error{s: s, err: e}
	}
	return Error{s: s, err: je}
}

// toStatus marshals the given jettison error into a *grpc.Status object,
// with a message given by the most recently wrapped error in the list of
// hops.
func toStatus(err error) *status.Status {
	s, ok := status.FromError(err)
	if !ok {
		c := codes.Unknown
		var msg string
		if errors.Is(err, context.Canceled) {
			c = codes.Canceled
		} else if errors.Is(err, context.DeadlineExceeded) {
			c = codes.DeadlineExceeded
		} else {
			msg = err.Error()
		}
		s = status.New(c, msg)
	}

	withWrap, err := s.WithDetails(errorToProto(err))
	if err != nil {
		log.Printf("jettison/errors: Failed to add WrappedError to status: %v", err)
	} else {
		s = withWrap
	}
	return s
}

// fromStatus will unmarshal a *grpc.Status into a jettison error object,
// returning a nil error if and only if no unexpected details were found on the
// status.
func fromStatus(s *status.Status) (*errors.JettisonError, bool) {
	if s == nil {
		return nil, false
	}
	for _, d := range s.Details() {
		if we, ok := d.(*jettisonpb.WrappedError); ok {
			return errorFromProto(we), true
		}
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
		we.Message = removeNonUTF8(je.Message)
		we.Binary = removeNonUTF8(je.Binary)
		we.Code = removeNonUTF8(je.Code)
		we.Source = removeNonUTF8(je.Source)
		if len(je.StackTrace) > 0 {
			we.StackTrace = make([]string, len(je.StackTrace))
			copy(we.StackTrace, je.StackTrace)
			for i := range we.StackTrace {
				we.StackTrace[i] = removeNonUTF8(we.StackTrace[i])
			}
		}
		we.KeyValues = kvToProto(je.KV)
	} else {
		we.Message = removeNonUTF8(err.Error())
	}
	switch unw := err.(type) {
	case interface{ Unwrap() error }:
		we.WrappedError = errorToProto(unw.Unwrap())
	case interface{ Unwrap() []error }:
		for _, e := range unw.Unwrap() {
			we.JoinedErrors = append(we.JoinedErrors, errorToProto(e))
		}
	}
	return &we
}

func kvToProto(kvs []models.KeyValue) []*jettisonpb.KeyValue {
	if len(kvs) == 0 {
		return nil
	}
	res := make([]*jettisonpb.KeyValue, 0, len(kvs))
	for _, kv := range kvs {
		res = append(res, &jettisonpb.KeyValue{
			Key:   removeNonUTF8(kv.Key),
			Value: removeNonUTF8(kv.Value),
		})
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

func removeNonUTF8(s string) string {
	return strings.ToValidUTF8(s, "[snip]")
}
