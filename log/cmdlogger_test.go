package log_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/sebdah/goldie/v2"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"github.com/luno/jettison/trace"
)

func TestCmdLogger(t *testing.T) {
	var buf bytes.Buffer
	log.SetCmdLoggerForTesting(t, &buf)

	ctx := log.ContextWith(context.TODO(), j.KS("ctx_key", "ctx_val"))
	log.Info(ctx, "this is an info message", j.KS("info_key", "info_val"))
	log.Error(ctx, io.EOF)
	log.Error(ctx, errors.New("example error"))

	goldie.New(t).Assert(t, "cmd_logger", trace.StripTestStacks(t, buf.Bytes()))
}
