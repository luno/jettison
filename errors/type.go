package errors

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"

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
	Source     string
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
	msg := "%s"
	args := []interface{}{je.Message}
	if p.Detail() && len(je.KV) > 0 {
		var fmts []string
		for _, kv := range je.KV {
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
	return je.Err
}

// Is returns true if the errors are equal as values, or the target is also
// a jettison error and je contains an error with the same code as the latest
// error in target. This is compatible with the Is interface from the Go 2
// error handling proposal.
func (je *JettisonError) Is(target error) bool {
	if je == nil {
		return target == nil
	}
	if je == target {
		return true
	}
	targetJErr, ok := target.(*JettisonError)
	if !ok {
		return false
	}
	if je.Code != "" {
		return targetJErr.Code == je.Code
	}
	// TODO(adam): Remove this behaviour
	if je.Message != "" {
		match := targetJErr.Message == je.Message
		if match {
			f := legacyCallback.Load()
			if f != nil {
				(*f)(je, target)
			}
		}
		return match
	}
	return false
}

var legacyCallback atomic.Pointer[func(src, target error)]

func SetLegacyCallback(f func(src, target error)) {
	legacyCallback.Store(&f)
}

func (je *JettisonError) IsZero() bool {
	return je.Message == "" &&
		je.Err == nil &&
		je.Binary == "" &&
		len(je.StackTrace) == 0 &&
		je.Code == "" &&
		je.Source == "" &&
		len(je.KV) == 0
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
