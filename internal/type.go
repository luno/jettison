package internal

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"golang.org/x/xerrors"

	"github.com/luno/jettison/models"
)

type Error struct {
	Message string
	Err     error

	Binary     string
	StackTrace []string
	Code       string
	Source     string
	KV         []models.KeyValue
}

// Format satisfies the fmt.Formatter interface providing customizable formatting:
//
//	%s, %v formats all wrapped error messages concatenated with ": ".
//	%+v, %#v does the above but also adds error parameters; "(k1=v1, k2=v2)".
func (je *Error) Format(state fmt.State, _ rune) {
	withParams := state.Flag(int('#')) || state.Flag(int('+'))
	p := &printer{Writer: state, detailed: withParams}
	var next xerrors.Formatter = je
	for {
		pre := p.written
		res := next.FormatError(p)
		if res == nil {
			return
		}
		if p.written > pre {
			_, _ = p.Write([]byte(": "))
		}
		formatter, ok := res.(xerrors.Formatter)
		if !ok {
			_, _ = p.Write([]byte(res.Error()))
			return
		}
		next = formatter
	}
}

// FormatError implements the Formatter interface for optionally detailed
// error message rendering - see the Go 2 error printing draft proposal for
// details.
func (je *Error) FormatError(p xerrors.Printer) error {
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
	return je.Err
}

func (je *Error) Error() string {
	return fmt.Sprintf("%v", je)
}

func (je *Error) String() string {
	return je.Error()
}

func (je *Error) Unwrap() error {
	return je.Err
}

// Is returns true if the errors are equal as values, or the target is also
// a jettison error and contains the same code as the target.
func (je *Error) Is(target error) bool {
	if je == nil {
		return target == nil
	}
	if je == target {
		return true
	}
	targetJErr, ok := target.(*Error)
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
			if f != nil && *f != nil {
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
	_ fmt.Formatter   = (*Error)(nil)
	_ xerrors.Printer = (*printer)(nil)
)
