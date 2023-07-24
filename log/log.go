package log

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/go-stack/stack"

	"github.com/luno/jettison/errors"
)

const (
	LevelInfo  Level = "info"
	LevelError Level = "error"
	LevelDebug Level = "debug"
)

type logOption func(*Entry)

func (o logOption) ApplyToLog(e *Entry) {
	o(e)
}

// WithLevel returns a jettison option to override the default log level.
// It only works when provided as option to log package functions.
func WithLevel(level Level) Option {
	return logOption(func(e *Entry) {
		e.Level = level
	})
}

// WithError returns a jettison option to add a structured error as part of
// Info logging. See Error for more details. It only works when provided
// as option to log package functions. Using this option while Error logging
// is not recommended.
func WithError(err error) Option {
	return logOption(func(e *Entry) {
		addErrorHops(e, err)

		// Add the most recent error code in the chain to the log's root.
		codes := errors.GetCodes(err)
		if len(codes) > 0 {
			e.ErrorCode = &codes[0]
		}
	})
}

type Option interface {
	ApplyToLog(*Entry)
}

func Debug(ctx context.Context, msg string, opts ...Option) {
	logger.Log(makeEntry(ctx, msg, LevelDebug, opts...))
}

// Info writes a structured jettison log to the logger. Any jettison
// key/value pairs contained in the given context are included in the log.
func Info(ctx context.Context, msg string, opts ...Option) {
	logger.Log(makeEntry(ctx, msg, LevelInfo, opts...))
}

// Error writes a structured jettison log of the given error to the logger.
// If the error is not already a Jettison error, it is converted into one and
// then logged. Any jettison key/value pairs contained in the given context are
// included in the log.
// If err is nil, a new error is created.
func Error(ctx context.Context, err error, opts ...Option) {
	if err == nil {
		err = errors.New("nil error logged - this is probably a bug")
	}
	opts = append(opts, WithError(err))
	e := makeEntry(ctx, err.Error(), LevelError, opts...)
	logger.Log(e)
}

func makeEntry(ctx context.Context, msg string, lvl Level, opts ...Option) Entry {
	l := newEntry(msg, lvl, 3)
	for _, o := range opts {
		o.ApplyToLog(&l)
	}
	l.Parameters = append(l.Parameters, ContextKeyValues(ctx)...)

	// Sort the parameters for consistent logging.
	sort.Slice(l.Parameters, func(i, j int) bool {
		return l.Parameters[i].Key < l.Parameters[j].Key
	})

	return l
}

// addErrorHops tries to convert the error to a jettison error
// and then adds the error as hops to the log. It also adds
// the error parameters log's root.
func addErrorHops(e *Entry, err error) {
	je, ok := err.(*errors.JettisonError)
	if !ok {
		je, ok = errors.New(err.Error(), errors.WithoutStackTrace()).(*errors.JettisonError)
	}
	if !ok {
		log.Printf("jettison/log: failed to convert error to jettison error: %v", err)
		// best-effort, will just log err.Err() wrapped as a Jettison log
	}

	for _, h := range je.Hops {
		e.Hops = append(e.Hops, h)
	}
	// Bubble up all nested parameters in the list of hops to the log's root
	// parameters list for ease of use. Newer parameters come first.
	for _, h := range je.Hops {
		for _, he := range h.Errors {
			if he.Parameters == nil {
				continue
			}

			for _, k := range he.Parameters {
				e.Parameters = append(e.Parameters, k)
			}
		}
	}
}

// newEntry returns an Entry struct decorated with useful defaults - stackSkip
// is the number of callstacks to skip in the stacktrace before pulling
// out the `source` of the call to `jettison/log.XXX`.
func newEntry(msg string, level Level, stackSkip int) Entry {
	return Entry{
		Message:   msg,
		Source:    fmt.Sprintf("%+v", stack.Caller(stackSkip)),
		Level:     level,
		Timestamp: time.Now(),
	}
}

type Interface interface {
	Debug(ctx context.Context, msg string, ol ...Option)
	Info(ctx context.Context, msg string, ol ...Option)
	Error(ctx context.Context, err error, ol ...Option)
}

type Jettison struct{}

func (j Jettison) Debug(ctx context.Context, msg string, ol ...Option) {
	Debug(ctx, msg, ol...)
}

func (j Jettison) Info(ctx context.Context, msg string, ol ...Option) {
	Info(ctx, msg, ol...)
}

func (j Jettison) Error(ctx context.Context, err error, ol ...Option) {
	Error(ctx, err, ol...)
}

var _ Interface = (*Jettison)(nil)
