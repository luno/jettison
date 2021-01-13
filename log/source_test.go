package log_test

import (
	"bytes"
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/log"
)

// TestSourceInfo tests the log source which includes line numbers.
// Adding anything to this file might break the test.
func TestSourceInfo(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetDefaultLoggerForTesting(t, buf)
	log.Info(nil, "message")
	verifyOutput(t, "source_info", buf.Bytes())
}

// TestSourceError tests the log source and stack trace which includes line numbers.
// Adding anything to this file might break the test.
func TestSourceError(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetDefaultLoggerForTesting(t, buf)
	log.Error(nil, errors.New("test error"))
	verifyOutput(t, "source_error", internal.StripTestStacks(t, buf.Bytes()))
}
