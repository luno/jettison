package log_test

import (
	"bytes"
	"context"
	stderrors "errors"
	"io"
	"testing"

	"github.com/sebdah/goldie/v2"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

//go:generate go test . -run TestCmdLogger -update

func TestCmdLogger(t *testing.T) {
	var buf bytes.Buffer
	log.SetCmdLoggerForTesting(t, &buf)
	errors.SetTraceConfigTesting(t, errors.TestingConfig)

	ctx := log.ContextWith(context.TODO(), j.KS("ctx_key", "ctx_val"))
	log.Info(ctx, "this is an info message", j.KS("info_key", "info_val"))
	log.Error(ctx, io.EOF)
	log.Error(ctx, errors.New("example error", j.KV("error_key", "error_val")))

	err := stderrors.Join(
		errors.New("error one"),
		errors.New("error two"),
	)
	log.Error(ctx, err)

	goldie.New(t).Assert(t, "cmd_logger", buf.Bytes())
}
