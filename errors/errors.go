package errors

import (
	"errors"

	"golang.org/x/xerrors"

	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
)

type ErrorOption func(je *JettisonError)

func (o ErrorOption) ApplyToError(je *JettisonError) {
	o(je)
}

// WithStackTrace will add a new stack trace to this error
func WithStackTrace() Option {
	bin, tr := getTrace(1)
	return ErrorOption(func(je *JettisonError) {
		je.Binary = bin
		je.StackTrace = tr
	})
}

// WithCode sets an error code on the error. A code should uniquely identity an error,
// the intention being to provide an equality check for jettison errors (see Is() for more details).
// The default code (the error message) doesn't provide strong unique guarantees.
func WithCode(code string) Option {
	return ErrorOption(func(je *JettisonError) {
		if len(je.Hops[0].Errors) > 0 {
			je.Hops[0].Errors[0].Code = code
		}
		je.Code = code
	})
}

// WithoutStackTrace clears any automatically populated stack trace.
// New always populates a stack trace and Wrap will if no sub error has a trace.
//
// This Option is useful for sentinel errors which have a useless init-time stack trace.
// Removing it allows a stacktrace to be added when it is Wrapped.
//
// Example
//
//	var ErrFoo = errors.New("foo", errors.WithoutStackTrace()) // Clear useless init-time stack trace.
//
//	func bar() error {
//	  return errors.Wrap(ErrFoo, "bar") // Wrapping ErrFoo adds a proper stack trace.
//	}
func WithoutStackTrace() Option {
	return ErrorOption(func(je *JettisonError) {
		if len(je.Hops[0].Errors) <= 1 {
			je.Hops[0].StackTrace = nil
		}
		je.Binary = ""
		je.StackTrace = nil
	})
}

func C(code string) Option {
	c := WithCode(code)
	st := WithoutStackTrace()
	return ErrorOption(func(je *JettisonError) {
		c.ApplyToError(je)
		st.ApplyToError(je)
	})
}

type Option interface {
	ApplyToError(je *JettisonError)
}

// New creates a new JettisonError with a populated stack trace
func New(msg string, ol ...Option) error {
	h := internal.NewHop()
	h.StackTrace = internal.GetStackTrace(2)
	h.Errors = []models.Error{
		internal.NewError(msg),
	}
	je := &JettisonError{
		Message: msg,
		Hops:    []models.Hop{h},
	}
	je.Binary, je.StackTrace = getTrace(1)
	for _, o := range ol {
		o.ApplyToError(je)
	}
	return je
}

// Wrap will wrap an existing error in a new JettisonError.
// If no error in the err error tree has a trace, a stack trace is populated.
func Wrap(err error, msg string, ol ...Option) error {
	if err == nil {
		return nil
	}

	// If err is a jettison error, we want to append to it's current segment's
	// list of errors. Othewise we want to just create a new Jettison error.
	je, ok := err.(*JettisonError)
	if !ok {
		je = &JettisonError{
			Hops:        []models.Hop{internal.NewHop()},
			OriginalErr: err,
		}

		je.Hops[0].Errors = []models.Error{
			{Message: err.Error()},
		}
	} else {
		// We don't want to mutate everyone's copy of the error.
		// When we only use nested errors, this stage will not be necessary, we'll always create a new struct
		je = je.Clone()
	}

	// If the current hop doesn't yet have a stack trace, add one.
	if je.Hops[0].StackTrace == nil {
		je.Hops[0].StackTrace = internal.GetStackTrace(2)
	}

	// Add the error to the stack and apply the options on the latest hop.
	je.Hops[0].Errors = append(
		[]models.Error{internal.NewError(msg)},
		je.Hops[0].Errors...,
	)

	je.Message = msg
	je.Err = err

	// We only need to add a trace when wrapping sentinel or non-jettison errors
	// for the first time
	if !hasTrace(err) {
		je.Binary, je.StackTrace = getTrace(1)
	}

	for _, o := range ol {
		o.ApplyToError(je)
	}
	return je
}

// Is is an alias of the standard library's errors.Is() function.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// IsAny returns true if Is(err, target) is true for any of the targets.
func IsAny(err error, targets ...error) bool {
	for _, target := range targets {
		if Is(err, target) {
			return true
		}
	}

	return false
}

// As is an alias of the standard library's errors.As() function.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Opaque is an alias of golang.org/x/exp/errors.Opaque().
// Deprecated. See https://github.com/golang/go/issues/29934#issuecomment-489682919
func Opaque(err error) error {
	return xerrors.Opaque(err)
}

// Unwrap is an alias of the standard library's errors.Unwrap() function.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// OriginalError returns the non-jettison error wrapped by the given one,
// if it exists. This is intended to provide limited interop with other error
// handling packages. This is best-effort - a jettison error that has been
// passed over the wire will no longer have an OriginalError().
func OriginalError(err error) error {
	jerr, ok := err.(*JettisonError)
	if !ok {
		return nil
	}

	return jerr.OriginalErr
}

// GetCodes returns the stack of error codes in the given jettison error chain.
// The error codes are returned in reverse-order of calls to Wrap(), i.e. the
// code of the latest wrapped error comes first in the list.
func GetCodes(err error) []string {
	je, ok := err.(*JettisonError)
	if !ok {
		return nil
	}

	var res []string
	for _, h := range je.Hops {
		for _, e := range h.Errors {
			if e.Code == "" {
				continue
			}

			res = append(res, e.Code)
		}
	}

	return res
}
