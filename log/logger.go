package log

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/luno/jettison/models"
)

// logger is the global logger. It defaults to a human friendly command line logger.
var logger Logger = newCmdLogger(os.Stderr, false)

// Log sub-types the internal log struct for the public interface.
type Log models.Log

// Logger does logging of log lines.
type Logger interface {
	// Log logs the given log and returns a string of what was written.
	Log(Log) string
}

// SetLogger sets the global logger.
func SetLogger(l Logger) {
	logger = l
}

func SetCmdLoggerForTesting(t testing.TB, w io.Writer) {
	cached := logger
	logger = newCmdLogger(w, true)

	t.Cleanup(func() {
		logger = cached
	})
}

func SetDefaultLoggerForTesting(t testing.TB, w io.Writer,
	opts ...Option,
) {
	cached := logger

	l := newJSONLogger(w, opts...)
	l.scrubTimestamp = true
	logger = l

	t.Cleanup(func() {
		logger = cached
	})
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

func (jl *jsonLogger) Log(l Log) string {
	il := models.Log(l)
	for _, o := range jl.opts {
		o.ApplyToLog(&il)
	}
	if jl.scrubTimestamp {
		il.Timestamp = time.Time{}
	}
	l = Log(il)

	res, err := json.Marshal(l)
	if err != nil {
		jl.logger.Printf("jettison/log: failed to marshal log: %v", err)
		jl.logger.Print(l.Message) // best-effort
		return l.Message
	}

	jl.logger.Print(string(res))
	return string(res)
}
