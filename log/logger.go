package log

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

// logger is the global logger. It defaults to a human friendly command line logger.
var logger Logger = NewCmdLogger(os.Stderr, false)

// Logger does logging of log lines.
type Logger interface {
	// Log logs the given log and returns a string of what was written.
	Log(context.Context, Entry) string
}

// SetLogger sets the global logger.
func SetLogger(l Logger) {
	logger = l
}

func SetLoggerForTesting(t testing.TB, l Logger) {
	old := logger
	t.Cleanup(func() {
		logger = old
	})
	logger = l
}

func SetCmdLoggerForTesting(t testing.TB, w io.Writer) {
	SetLoggerForTesting(t, NewCmdLogger(w, true))
}

func SetDefaultLoggerForTesting(t testing.TB, w io.Writer, opts ...Option) {
	l := newJSONLogger(w, opts...)
	l.scrubTimestamp = true

	SetLoggerForTesting(t, l)
}

func newJSONLogger(w io.Writer, opts ...Option) *jsonLogger {
	return &jsonLogger{
		logger: log.New(w, "", 0),
		opts:   opts,
	}
}

// jsonLogger is the default logger which writes json to stdout.
type jsonLogger struct {
	logger *log.Logger

	// default options and other flags for testing
	opts           []Option
	scrubTimestamp bool
}

func (jl *jsonLogger) Log(_ context.Context, l Entry) string {
	for _, o := range jl.opts {
		o.ApplyToLog(&l)
	}
	if jl.scrubTimestamp {
		l.Timestamp = time.Time{}
	}

	res, err := json.Marshal(l)
	if err != nil {
		jl.logger.Printf("jettison/log: failed to marshal log: %v", err)
		jl.logger.Print(l.Message) // best-effort
		return l.Message
	}

	jl.logger.Print(string(res))
	return string(res)
}
