package log_test

import (
	"bytes"
	"testing"

	jlog "github.com/luno/jettison/log"
)

// TestSource tests the log source which includes line numbers.
// Adding anything to this file might break the test.
func TestSource(t *testing.T) {
	buf := new(bytes.Buffer)
	jlog.SetDefaultLoggerForTesting(t, buf)
	jlog.Info(nil, "message")
	verifyOutput(t, "source", buf.Bytes())
}
