package log_test

import (
	"bytes"
	"context"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"io"
	"testing"
)

func TestCmdLogger(t *testing.T) {
	var buf bytes.Buffer
	log.SetCmdLoggerForTesting(t, &buf)

	ctx := log.ContextWith(context.TODO(), j.KS("ctx_key", "ctx_val"))
	log.Info(ctx, "this is an info message", j.KS("info_key", "info_val"))
	log.Error(ctx, io.EOF)
	log.Error(ctx, errors.New("example error"))

	verifyOutput(t, "cmd_logger", buf.Bytes())
}
