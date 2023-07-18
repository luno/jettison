package errors

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"

	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
)

type errorOption func(je *JettisonError)

func (o errorOption) ApplyToError(je *JettisonError) {
	o(je)
}

// WithBinary sets the binary of the current hop to the given value.
func WithBinary(bin string) Option {
	return errorOption(func(je *JettisonError) {
		je.Hops[0].Binary = bin
		je.Metadata.Trace.Binary = bin
	})
}

// WithCode sets an error code on the latest error in the chain. A code should
// uniquely identity an error, the intention being to provide a notion of
// equality for jettison errors (see Is() for more details).
// Note the default code (error message) doesn't provide strong unique guarantees.
func WithCode(code string) Option {
	return errorOption(func(je *JettisonError) {
		if len(je.Hops[0].Errors) > 0 {
			je.Hops[0].Errors[0].Code = code
		}
		je.Metadata.Code = code
	})
}

// WithoutStackTrace clears the stacktrace if this is the first
// error in the chain. This is useful for sentinel errors
// with useless init-time stacktrace allowing a proper
// stacktrace to be added when wrapping them.
//
// Example
//
//	var ErrFoo = errors.New("foo", errors.WithoutStackTrace()) // Clear useless init-time stack trace.
//
//	func bar() error {
//	  return errors.Wrap(ErrFoo, "bar") // Wrapping ErrFoo adds a proper stack trace.
//	}
func WithoutStackTrace() Option {
	return errorOption(func(je *JettisonError) {
		if len(je.Hops[0].Errors) <= 1 {
			je.Hops[0].StackTrace = nil
		}
		je.Metadata.Trace = models.Hop{}
	})
}

func C(code string) Option {
	c := WithCode(code)
	st := WithoutStackTrace()
	return errorOption(func(je *JettisonError) {
		c.ApplyToError(je)
		st.ApplyToError(je)
	})
}

type Option interface {
	ApplyToError(je *JettisonError)
}

func New(msg string, ol ...Option) error {
	h := internal.NewHop()
	h.StackTrace = internal.GetStackTrace(2)
	h.Errors = []models.Error{
		internal.NewError(msg),
	}
	md := newMetadata()
	je := &JettisonError{
		Message:  msg,
		Metadata: md,
		Hops:     []models.Hop{h},
	}
	for _, o := range ol {
		o.ApplyToError(je)
	}
	return je
}

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

	var md models.Metadata
	// We only need to add a trace when wrapping sentinel or non-jettison errors
	// for the first time
	if _, has := hasTrace(err); !has {
		md.Trace = trace()
	}

	// For the nested wrapping we're only interested in wrapping this error message,
	// so we can overwrite the nested fields.
	je.Message = msg
	je.Err = err
	je.Metadata = md

	for _, o := range ol {
		o.ApplyToError(je)
	}
	return je
}

func newMetadata() models.Metadata {
	return models.Metadata{
		Trace: models.Hop{
			Binary:     filepath.Base(os.Args[0]),
			StackTrace: internal.GetStackTrace(3),
		},
	}
}

func trace() models.Hop {
	return models.Hop{
		Binary: filepath.Base(os.Args[0]),
		// Skip GetStackTrace, trace, and New/Wrap
		StackTrace: internal.GetStackTrace(3),
	}
}

type unwrapper interface {
	Unwrap() error
}

type unwrapperList interface {
	Unwrap() []error
}

func hasTrace(err error) (models.Hop, bool) {
	e := err
	for e != nil {
		if je, ok := e.(*JettisonError); ok && je.Metadata.Trace.Binary != "" {
			return je.Metadata.Trace, true
		}
		if un, ok := e.(unwrapper); ok {
			e = un.Unwrap()
		} else {
			break
		}
	}
	return models.Hop{}, false
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
