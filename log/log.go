package log

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/go-stack/stack"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
)

const (
	LevelInfo  models.Level = "info"
	LevelError models.Level = "error"
	LevelDebug models.Level = "debug"
)

type logOption func(*models.Log)

func (o logOption) ApplyToLog(l *models.Log) {
	o(l)
}

// WithLevel returns a jettison option to override the default log level.
// It only works when provided as option to log package functions.
func WithLevel(level models.Level) Option {
	return logOption(func(l *models.Log) {
		l.Level = level
	})
}

// WithError returns a jettison option to add a structured error as part of
// Info logging. See Error for more details. It only works when provided
// as option to log package functions. Using this option while Error logging
// is not recommended.
func WithError(err error) Option {
	return logOption(func(l *models.Log) {
		addErrorHops(l, err)

		// Add the most recent error code in the chain to the log's root.
		codes := errors.GetCodes(err)
		if len(codes) > 0 {
			l.ErrorCode = &codes[0]
		}
	})
}

// WithStackTrace returns a jettison option to add a stacktrace as a hop to the log.
// It only works when provided as option to log package functions.
func WithStackTrace() Option {
	return logOption(func(l *models.Log) {
		h := internal.NewHop()
		h.StackTrace = internal.GetStackTrace(2)
		l.Hops = append(l.Hops, h)
	})
}

type Option interface {
	ApplyToLog(*models.Log)
}

func Debug(ctx context.Context, msg string, opts ...Option) {
	l := makeLog(ctx, msg, LevelDebug, opts...)
	logger.Log(Log(l))
}

// Info writes a structured jettison log to the logger. Any jettison
// key/value pairs contained in the given context are included in the log.
func Info(ctx context.Context, msg string, opts ...Option) {
	l := makeLog(ctx, msg, LevelInfo, opts...)
	logger.Log(Log(l))
}

// Error writes a structured jettison log of the given error to the logger.
// If the error is not already a Jettison error, it is converted into one and
// then logged. Any jettison key/value pairs contained in the given context are
// included in the log.
// If err is nil, a new error is created.
func Error(ctx context.Context, err error, opts ...Option) {
	if err != nil {
		opts = append(opts, WithError(err))
	} else {
		err = errors.New("nil error logged - this is probably a bug")
		opts = append(opts, WithStackTrace())
	}
	l := makeLog(ctx, err.Error(), LevelError, opts...)
	logger.Log(Log(l))
}

func makeLog(ctx context.Context, msg string, lvl models.Level, opts ...Option) models.Log {
	l := newLog(msg, lvl, 3)
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
func addErrorHops(l *models.Log, err error) {
	je, ok := err.(*errors.JettisonError)
	if !ok {
		je, ok = errors.New(err.Error(), errors.WithoutStackTrace()).(*errors.JettisonError)
	}
	if !ok {
		log.Printf("jettison/log: failed to convert error to jettison error: %v", err)
		// best-effort, will just log err.Err() wrapped as a Jettison log
	}

	for _, h := range je.Hops {
		l.Hops = append(l.Hops, h)
	}
	// Bubble up all nested parameters in the list of hops to the log's root
	// parameters list for ease of use. Newer parameters come first.
	for _, h := range je.Hops {
		for _, e := range h.Errors {
			if e.Parameters == nil {
				continue
			}

			for _, k := range e.Parameters {
				l.Parameters = append(l.Parameters, k)
			}
		}
	}
}

// newLog returns a Log struct decorated with useful defaults - stackSkip
// is the number of callstacks to skip in the stacktrace before pulling
// out the `source` of the call to `jettison/log.XXX`.
func newLog(msg string, level models.Level, stackSkip int) models.Log {
	return models.Log{
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
