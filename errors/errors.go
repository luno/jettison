package errors

import (
	"github.com/luno/jettison"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
	"golang.org/x/xerrors"
)

// WithBinary sets the binary of the current hop to the given value.
func WithBinary(bin string) jettison.OptionFunc {
	return func(d jettison.Details) {
		h, ok := d.(*models.Hop)
		if !ok {
			return
		}

		h.Binary = bin
	}
}

// WithCode sets an error code on the latest error in the chain. A code should
// uniquely identity an error, the intention being to provide a notion of
// equality for jettison errors (see Is() for more details).
// Note the default code (error message) doesn't provide strong unique guarantees.
func WithCode(code string) jettison.OptionFunc {
	return func(d jettison.Details) {
		h, ok := d.(*models.Hop)
		if !ok || len(h.Errors) == 0 {
			return
		}

		h.Errors[0].Code = code
	}
}

// WithoutStackTrace clears the stacktrace if this is the first
// error in the chain. This is useful for sentinel errors
// with useless init-time stacktrace allowing a proper
// stacktrace to be added when wrapping them.
//
// Example
//  var ErrFoo = errors.New("foo", errors.WithoutStackTrace()) // Clear useless init-time stack trace.
//
//  func bar() error {
//    return errors.Wrap(ErrFoo, "bar") // Wrapping ErrFoo adds a proper stack trace.
//  }
func WithoutStackTrace() jettison.OptionFunc {
	return func(d jettison.Details) {
		h, ok := d.(*models.Hop)
		if !ok || len(h.Errors) > 1 {
			return
		}

		h.StackTrace = nil
	}
}

func New(msg string, ol ...jettison.Option) error {
	h := internal.NewHop()
	h.StackTrace = internal.GetStackTrace(2)
	h.Errors = []models.Error{
		internal.NewError(msg),
	}

	for _, o := range ol {
		o.Apply(&h)
	}

	return &JettisonError{
		Hops: []models.Hop{h},
	}
}

func Wrap(err error, msg string, ol ...jettison.Option) error {
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

	for _, o := range ol {
		o.Apply(&je.Hops[0])
	}

	return je
}

// Is is an alias of golang.org/x/errors.Is()
func Is(err, target error) bool {
	return xerrors.Is(err, target)
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

// As is an alias of golang.org/x/exp/errors.As().
func As(err error, target interface{}) bool {
	return xerrors.As(err, target)
}

// Opaque is an alias of golang.org/x/exp/errors.Opaque().
func Opaque(err error) error {
	return xerrors.Opaque(err)
}

// Unwrap is an alias of golang.org/x/exp/errors.Unwrap().
func Unwrap(err error) error {
	return xerrors.Unwrap(err)
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
