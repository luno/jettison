package errors

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/internal/jettisonpb"
	"github.com/luno/jettison/models"
)

var (
	defaultCode = flag.Int("jettison_default_error_code",
		int(codes.Unknown), "Default error code; see google.golang.org/grpc/codes")

	ErrInvalidError = errors.New("jettison/errors: given grpc.Status does not contain a valid jettison error")
)

// WithStackTrace sets the stack trace of the current hop to the given value.
func WithStackTrace(trace []string) jettison.OptionFunc {
	return func(d jettison.Details) {
		h, ok := d.(*models.Hop)
		if !ok {
			return
		}

		h.StackTrace = trace
	}
}

// JettisonError is the internal error representation. We use a separate type
// so that we can implement the Go 2.0 error interfaces, and we also need to
// implement the GRPCStatus() method so that jettison errors can be passed over
// gRPC seamlessly.
//
// See https://github.com/grpc/grpc-go/blob/master/status/status.go#L130.
type JettisonError struct {
	Hops []models.Hop

	// If we've wrapped a non-Jettison error, we lose interop with other error
	// libraries such as github.com/pkg/errors. This is a best-effort attempt
	// to keep this interop - err.Cause() will return originalErr.
	OriginalErr error
}

// GRPCStatus marshals the given jettison error into a *grpc.Status object,
// with a message given by the most recently wrapped error in the list of
// hops.
func (je *JettisonError) GRPCStatus() *status.Status {
	msg := ""
	if le, ok := je.LatestError(); ok {
		msg = le.Message
	}

	c := getDefaultCode()
	switch je.OriginalErr {
	case context.Canceled:
		c = codes.Canceled
	case context.DeadlineExceeded:
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

	return res
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
	je = je.Clone() // don't want to modify the reference

	for len(je.Hops) > 0 {
		if len(je.Hops[0].Errors) == 0 {
			je.Hops = je.Hops[1:]
			continue
		}

		je.Hops[0].Errors = je.Hops[0].Errors[1:]
		break
	}

	// Remove any empty hop layers.
	for len(je.Hops) > 0 {
		h := je.Hops[0]
		if len(h.Errors) > 0 {
			break
		}

		je.Hops = je.Hops[1:]
	}

	if len(je.Hops) == 0 {
		return nil
	}

	// If this was the last unwrap, just return the original error for
	// compatibility with golang.org/x/xerrors.Unwrap() if it isn't nil.
	//
	// NOTE(guy): This can lead to a gotcha where an errors.Is() call works
	// locally, but doesn't work over gRPC since original error values can't
	// be preserved over the wire.
	if len(je.Hops) == 1 &&
		len(je.Hops[0].Errors) == 1 &&
		je.OriginalErr != nil {

		return je.OriginalErr
	}

	return je
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

// As finds the first error in the jettison error chain that matches target's
// type. This works with non-jettison errors that have been wrapped by
// jettison so long as the error hasn't been passed over gRPC.
// Note: target MUST be a pointer to a non-nil error value, or As will panic.
func (je *JettisonError) As(target interface{}) bool {
	if target == nil {
		panic("jettison/errors: target cannot be nil")
	}

	typ := reflect.TypeOf(target)
	if typ.Kind() != reflect.Ptr {
		panic("jettison/errors: target must be a pointer")
	}

	var iface error
	if !typ.Elem().Implements(reflect.TypeOf(&iface).Elem()) {
		panic("jettison/errors: target must be a pointer to an error type")
	}

	// If target is a pointer to a jettison error, we can just set its value.
	_, ok := target.(**JettisonError)
	if ok {
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(je))
		return true
	}

	// Otherwise, we need to check the jettison error's OriginalErr.
	for je != nil {
		if je.OriginalErr != nil && reflect.TypeOf(je.OriginalErr) == typ.Elem() {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(je.OriginalErr))
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

// FromStatus unmarshals a *grpc.Status into a jettison error object,
// returning a nil error if and only if no unexpected details were found on the
// status.
func FromStatus(s *status.Status) (*JettisonError, error) {
	if s == nil {
		return nil, ErrInvalidError
	} else if len(s.Details()) == 0 {
		return nil, ErrInvalidError
	}

	var res JettisonError
	for _, d := range s.Details() {
		spb, ok := d.(*jettisonpb.Hop)
		if !ok {
			return nil, ErrInvalidError
		}

		s, err := internal.HopFromProto(spb)
		if err != nil {
			return nil, err
		}

		res.Hops = append(res.Hops, *s)
	}

	return &res, nil
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

func getDefaultCode() codes.Code {
	return codes.Code(*defaultCode)
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

var _ fmt.Formatter = (*JettisonError)(nil)
var _ xerrors.Printer = (*printer)(nil)
