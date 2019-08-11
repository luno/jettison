package log

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/luno/jettison"
	"github.com/luno/jettison/models"
)

var logger Logger = newJSONLogger(os.Stdout)

// Log sub-types the internal log struct for the public interface.
type Log models.Log

type Logger interface {
	// Log logs the given log and returns a string of what was written.
	Log(Log) string
}

// SetLogger sets the global logger.
func SetLogger(l Logger) {
	logger = l
}

func SetDefaultLoggerForTesting(_ testing.TB, w io.Writer,
	opts ...jettison.Option) {

	l := newJSONLogger(w, opts...)
	l.scrubTimestamp = true
	logger = l
}

func newJSONLogger(w io.Writer, opts ...jettison.Option) *jsonLogger {
	return &jsonLogger{
		logger: log.New(w, "", 0),
		opts:   opts,
	}
}

// jsonLogger is the default logger which writes json to stdout.
type jsonLogger struct {
	logger *log.Logger

	// default options and other flags for testing
	opts           []jettison.Option
	scrubTimestamp bool
}

func (jl *jsonLogger) Log(l Log) string {
	il := models.Log(l)
	for _, o := range jl.opts {
		o.Apply(&il)
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
