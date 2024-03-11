package errors

import (
	stderrors "errors"

	"github.com/luno/jettison/internal"
)

type ErrorOption func(je *internal.Error)

func (o ErrorOption) ApplyToError(je *internal.Error) {
	o(je)
}

// WithStackTrace will add a new stack trace to this error
func WithStackTrace() Option {
	bin, tr := getTrace(1)
	return ErrorOption(func(je *internal.Error) {
		je.Binary = bin
		je.StackTrace = tr
	})
}

// WithCode sets an error code on the error. A code should uniquely identity an error,
// the intention being to provide an equality check for jettison errors (see Is() for more details).
// The default code (the error message) doesn't provide strong unique guarantees.
func WithCode(code string) Option {
	return ErrorOption(func(je *internal.Error) {
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
	return ErrorOption(func(je *internal.Error) {
		je.Binary = ""
		je.StackTrace = nil
	})
}

func C(code string) Option {
	c := WithCode(code)
	st := WithoutStackTrace()
	return ErrorOption(func(je *internal.Error) {
		c.ApplyToError(je)
		st.ApplyToError(je)
	})
}

type Option interface {
	ApplyToError(je *internal.Error)
}

// New creates a new JettisonError with a populated stack trace
func New(msg string, ol ...Option) error {
	je := &internal.Error{
		Message: msg,
		Source:  getSourceCode(1),
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
	je := &internal.Error{
		Message: msg,
		Err:     err,
		Source:  getSourceCode(1),
	}
	// We only need to add a trace when wrapping sentinel or non-jettison errors
	// for the first time
	if _, _, found := GetLastStackTrace(err); !found {
		je.Binary, je.StackTrace = getTrace(1)
	}
	for _, o := range ol {
		o.ApplyToError(je)
	}
	return je
}

// Is is an alias of the standard library's errors.Is() function.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
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
func As(err error, target any) bool {
	return stderrors.As(err, target)
}

// Unwrap is an alias of the standard library's errors.Unwrap() function.
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

// Join is an alias of the standard library's errors.Join() function.
func Join(err ...error) error {
	return stderrors.Join(err...)
}

// GetCodes returns the stack of error codes in the given jettison error chain.
// The error codes are returned in reverse-order of calls to Wrap(), i.e. the
// code of the latest wrapped error comes first in the list.
func GetCodes(err error) []string {
	var ret []string
	Walk(err, func(err error) bool {
		je, ok := err.(*internal.Error)
		if !ok {
			return true
		}
		if je.Code != "" {
			ret = append(ret, je.Code)
		} else if je.Message != "" {
			// TODO(adam): Remove this behaviour, we shouldn't use message strings as codes
			ret = append(ret, je.Message)
		}
		return true
	})
	return ret
}

func GetLastStackTrace(err error) (string, []string, bool) {
	var bin string
	var stack []string
	var found bool
	Walk(err, func(err error) bool {
		je, ok := err.(*internal.Error)
		if !ok || je.Binary == "" {
			return true
		}
		bin = je.Binary
		stack = je.StackTrace
		found = true
		return false
	})
	return bin, stack, found
}

// GetKeyValues returns all embedded key value info in the error
func GetKeyValues(err error) map[string]string {
	ret := make(map[string]string)
	Walk(err, func(err error) bool {
		je, ok := err.(*internal.Error)
		if ok {
			for _, kv := range je.KV {
				if _, ok := ret[kv.Key]; ok {
					continue
				}
				ret[kv.Key] = kv.Value
			}
		}
		return true
	})
	return ret
}

// Walk will do a depth first traversal of the error tree.
// do is called for each error on the traversal, if it returns false,
// then the traversal will be terminated
func Walk(err error, do func(error) bool) {
	walkRecur(err, do)
}

func walkRecur(err error, do func(error) bool) bool {
	for err != nil {
		if !do(err) {
			return false
		}
		switch unw := err.(type) {
		case interface{ Unwrap() error }:
			err = unw.Unwrap()
			if err == nil {
				return true
			}
		case interface{ Unwrap() []error }:
			for _, e := range unw.Unwrap() {
				if !walkRecur(e, do) {
					return false
				}
			}
			return true
		default:
			return true
		}
	}
	return true
}

// Flatten walks the error tree, creating a path for each leaf of the tree
// if the tree looks like this:
//
//	 ── a
//		└── b
//		    ├── c
//		    │   └── e
//		    │       └── f
//		    └── d
//		        ├── g
//		        └── h
//
// Then the paths we will get are:
// [a, b, c, e, f]
// [a, b, d, g]
// [a, b, d, h]
func Flatten(err error) [][]error {
	var ret [][]error
	paths := [][]error{{err}}
	for len(paths) > 0 {
		p := paths[0]
		paths = paths[1:]

		nxt, ok := extendPath(p)
		if ok {
			paths = append(nxt, paths...)
		} else {
			ret = append(ret, p)
		}
	}
	return ret
}

func extendPath(path []error) ([][]error, bool) {
	if len(path) == 0 {
		return nil, false
	}
	last := path[len(path)-1]
	switch unw := last.(type) {
	case interface{ Unwrap() error }:
		nxt := unw.Unwrap()
		if nxt != nil {
			return [][]error{append(path, nxt)}, true
		}
	case interface{ Unwrap() []error }:
		var ret [][]error
		for _, nxt := range unw.Unwrap() {
			p := make([]error, len(path), len(path)+1)
			copy(p, path)
			p = append(p, nxt)
			ret = append(ret, p)
		}
		return ret, true
	}
	return nil, false
}
