package errors

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/xerrors"

	"github.com/luno/jettison/models"
)

// JettisonError is the internal error representation. We use a separate type
// so that we can implement the Go 2.0 error interfaces, and we also need to
// implement the GRPCStatus() method so that jettison errors can be passed over
// gRPC seamlessly.
//
// See https://github.com/grpc/grpc-go/blob/master/status/status.go#L130.
type JettisonError struct {
	Message string
	Err     error

	Binary     string
	StackTrace []string
	Code       string
	KV         []models.KeyValue

	Hops []models.Hop

	// If we've wrapped a non-Jettison error, we lose interop with other error
	// libraries such as github.com/pkg/errors. This is a best-effort attempt
	// to keep this interop - err.Cause() will return originalErr.
	OriginalErr error
}

// Format satisfies the fmt.Formatter interface providing customizable formatting:
//
//	%s, %v formats all wrapped error messages concatenated with ": ".
//	%+v, %#v does the above but also adds error parameters; "(k1=v1, k2=v2)".
func (je *JettisonError) Format(state fmt.State, c rune) {
	withParams := state.Flag(int('#')) || state.Flag(int('+'))
	p := &printer{Writer: state, detailed: withParams}
	next := je
	for {
		pre := p.written
		res := next.FormatError(p)
		if res == nil {
			return
		}
		if p.written > pre {
			_, _ = p.Write([]byte(": "))
		}

		jerr, ok := res.(*JettisonError)
		if !ok {
			_, _ = p.Write([]byte(res.Error()))
			return
		}
		next = jerr
	}
}

// FormatError implements the Formatter interface for optionally detailed
// error message rendering - see the Go 2 error printing draft proposal for
// details.
func (je *JettisonError) FormatError(p xerrors.Printer) error {
	le, ok := je.LatestError()
	if !ok {
		return nil
	}

	msg := "%s"
	args := []interface{}{le.Message}
	if p.Detail() && len(le.Parameters) > 0 {
		var fmts []string
		for _, kv := range le.Parameters {
			fmts = append(fmts, "%s")
			args = append(args, kv.Key+"="+kv.Value)
		}

		msg += "(" + strings.Join(fmts, ", ") + ")"
	}

	p.Printf(msg, args...)
	return je.Unwrap()
}

// Error satisfies the built-in error interface and returns the default error format.
func (je *JettisonError) Error() string {
	return fmt.Sprintf("%v", je)
}

func (je *JettisonError) String() string {
	return je.Error()
}

// LatestError returns the error that was added last to the given jettison
// error, or false if there isn't one.
func (je *JettisonError) LatestError() (models.Error, bool) {
	for _, h := range je.Hops {
		for _, e := range h.Errors {
			return e, true
		}
	}

	return models.Error{}, false
}

// Clone returns a copy of the jettison error that can be safely mutated.
func (je *JettisonError) Clone() *JettisonError {
	res := JettisonError{
		Message:     je.Message,
		OriginalErr: je.OriginalErr,
	}

	for _, h := range je.Hops {
		res.Hops = append(res.Hops, h.Clone())
	}

	return &res
}

// Unwrap returns the next error in the jettison error chain, or nil if there
// is none. This is compatible with the Wrapper interface from the Go 2 error
// inspection proposal.
func (je *JettisonError) Unwrap() error {
	err := je.Clone() // Don't want to modify the reference

	for len(err.Hops) > 0 {
		if len(err.Hops[0].Errors) == 0 {
			err.Hops = err.Hops[1:]
			continue
		}

		err.Hops[0].Errors = err.Hops[0].Errors[1:]
		break
	}

	// Remove any empty hop layers.
	for len(err.Hops) > 0 {
		h := err.Hops[0]
		if len(h.Errors) > 0 {
			break
		}

		err.Hops = err.Hops[1:]
	}

	if len(err.Hops) == 0 {
		return nil
	}

	// If this was the last unwrap, just return the original error for
	// compatibility with golang.org/x/xerrors.Unwrap() if it isn't nil.
	//
	// NOTE(guy): This can lead to a gotcha where an errors.Is() call works
	// locally, but doesn't work over gRPC since original error values can't
	// be preserved over the wire.
	if len(err.Hops) == 1 &&
		len(err.Hops[0].Errors) == 1 &&
		err.OriginalErr != nil {

		return err.OriginalErr
	}

	if subJe, ok := je.Err.(*JettisonError); ok {
		err.Message = subJe.Message
		err.Binary = subJe.Binary
		err.StackTrace = subJe.StackTrace
		err.KV = subJe.KV
		err.Code = subJe.Code
		err.Err = subJe.Err
	}

	return err
}

// Is returns true if the errors are equal as values, or the target is also
// a jettison error and je contains an error with the same code as the latest
// error in target. This is compatible with the Is interface from the Go 2
// error handling proposal.
func (je *JettisonError) Is(target error) bool {
	if je == nil || target == nil {
		return false
	}

	targetJErr, ok := target.(*JettisonError)
	if !ok {
		return false
	}

	if je == targetJErr {
		return true
	}

	targetErr, ok := targetJErr.LatestError()
	if !ok {
		return false
	}

	// If target doesn't have an error code, then there's nothing to check.
	if targetErr.Code == "" {
		return false
	}

	for je != nil {
		jerr, ok := je.LatestError()
		if !ok {
			return false
		}

		if targetErr.Code == jerr.Code {
			return true
		}

		je = unwrap(je)
	}

	return false
}

// GetKey returns the value of the first jettison key/value pair with the
// given key in the error chain.
func (je *JettisonError) GetKey(key string) (string, bool) {
	for _, h := range je.Hops {
		for _, e := range h.Errors {
			for _, p := range e.Parameters {
				if p.Key == key {
					return p.Value, true
				}
			}
		}
	}

	return "", false
}

// GetKeyValues returns all the jettison keys in the error chain.
// Note that if two errors have the same key, only the earliest will be stored.
func (je *JettisonError) GetKeyValues() map[string]string {
	keyValues := make(map[string]string)
	for _, h := range je.Hops {
		for _, e := range h.Errors {
			for _, p := range e.Parameters {
				keyValues[p.Key] = p.Value
			}
		}
	}

	return keyValues
}

func (je *JettisonError) IsZero() bool {
	return je.Message == "" &&
		je.Err == nil &&
		je.Binary == "" &&
		len(je.StackTrace) == 0 &&
		je.Code == "" &&
		len(je.KV) == 0
}

// unwrap is a thin wrapper around internal.JettisonError that returns another
// internal.JettisonError instead of an error interface value.
func unwrap(je *JettisonError) *JettisonError {
	if je == nil {
		return nil
	}

	err := je.Unwrap()
	res, ok := err.(*JettisonError)
	if !ok {
		return nil
	}
	return res
}

// printer implements xerrors.Printer interface.
type printer struct {
	io.Writer
	detailed bool
	written  int
}

func (p *printer) Print(args ...interface{}) {
	w, _ := p.Write([]byte(fmt.Sprint(args...)))
	p.written += w
}

func (p *printer) Printf(format string, args ...interface{}) {
	w, _ := p.Write([]byte(fmt.Sprintf(format, args...)))
	p.written += w
}

func (p *printer) Detail() bool {
	return p.detailed
}

var (
	_ fmt.Formatter   = (*JettisonError)(nil)
	_ xerrors.Printer = (*printer)(nil)
)
