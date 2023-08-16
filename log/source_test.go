package log_test

import (
	"bytes"
	"testing"

	"github.com/sebdah/goldie/v2"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/log"
	"github.com/luno/jettison/trace"
)

// TestSourceInfo tests the log source which includes line numbers.
// Adding anything to this file might break the test.
func TestSourceInfo(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetDefaultLoggerForTesting(t, buf)
	log.Info(nil, "message")
	goldie.New(t).Assert(t, "source_info", buf.Bytes())
}

// TestSourceError tests the log source and stack trace which includes line numbers.
// Adding anything to this file might break the test.
func TestSourceError(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetDefaultLoggerForTesting(t, buf)
	log.Error(nil, errors.New("test error"))
	goldie.New(t).Assert(t, "source_error", trace.StripTestStacks(t, buf.Bytes()))
}
