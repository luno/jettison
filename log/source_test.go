package log_test

import (
	"bytes"
	"testing"

	"github.com/sebdah/goldie/v2"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/log"
)

// TestSourceInfo tests the log source which includes line numbers.
// Adding anything to this file might break the test.
func TestSourceInfo(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetDefaultLoggerForTesting(t, buf)
	errors.SetTraceConfigTesting(t, errors.TestingConfig)
	log.Info(nil, "message")
	goldie.New(t).Assert(t, "source_info", buf.Bytes())
}

// TestSourceError tests the log source and stack trace which includes line numbers.
// Adding anything to this file might break the test.
func TestSourceError(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetDefaultLoggerForTesting(t, buf)
	errors.SetTraceConfigTesting(t, errors.TestingConfig)
	log.Error(nil, errors.New("test error"))
	goldie.New(t).Assert(t, "source_error", buf.Bytes())
}
