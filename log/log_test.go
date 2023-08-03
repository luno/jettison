package log_test

import (
	"bytes"
	"context"
	"flag"
	stdlib_log "log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jerrors "github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	jlog "github.com/luno/jettison/log"
)

var writeGoldenFiles = flag.Bool("write-golden-files", false,
	"Whether or not to overwrite golden files with test output.")

//go:generate go test . -write-golden-files

type source string

func (s source) ApplyToLog(e *jlog.Entry) {
	e.Source = string(s)
}

func (s source) ApplyToError(je *jerrors.JettisonError) {
	je.Hops[0].SetSource(string(s))
}

// WithCustomTrace sets the stack trace of the current hop to the given value.
func WithCustomTrace(bin string, stack []string) jerrors.ErrorOption {
	return func(je *jerrors.JettisonError) {
		je.Hops[0].Binary = bin
		je.Hops[0].StackTrace = stack
		je.Binary = bin
		je.StackTrace = stack
	}
}

func TestLog(t *testing.T) {
	testCases := []struct {
		name string
		msg  string
		ctx  context.Context
		opts []jlog.Option
	}{
		{
			name: "message_only",
			msg:  "test_message",
		},
		{
			name: "message_with_kv",
			msg:  "test_message",
			opts: []jlog.Option{
				j.KV("key", "value"),
			},
		},
		{
			name: "message_with_error_level",
			msg:  "test_message",
			opts: []jlog.Option{
				jlog.WithLevel(jlog.LevelError),
			},
		},
		{
			name: "message_with_unordered_parameters",
			msg:  "test_message",
			opts: []jlog.Option{
				j.KV("a", "c"),
				j.KV("c", "d"),
				j.KV("d", "c"),
				j.KV("c", "a"),
			},
		},
		{
			name: "message_with_error",
			msg:  "test_message",
			opts: []jlog.Option{
				jlog.WithError(jerrors.New("test",
					source("testsource"),
					WithCustomTrace("testservice", []string{"teststacktrace"}),
				)),
			},
		},
		{
			name: "message_with_context",
			ctx:  jlog.ContextWith(context.Background(), j.KS("ctx_key", "ctx_val")),
			msg:  "test_message",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			jlog.SetDefaultLoggerForTesting(t, buf, source("testsource"))
			jlog.Info(tc.ctx, tc.msg, tc.opts...)

			verifyOutput(t, "log_"+tc.name, buf.Bytes())
		})
	}
}

func TestError(t *testing.T) {
	// TODO(adam): Fix this test, writes different stacktrace on amd/arm hardware
	t.Skip("skipped due to non-deterministic logging details")
	testCases := []struct {
		name string
		ctx  context.Context
		err  error
	}{
		{
			name: "nil_error",
			err:  nil,
		},
		{
			name: "message_only",
			err: jerrors.New("test",
				source("testsource"),
				WithCustomTrace("testservice", []string{"teststacktrace"}),
			),
		},
		{
			name: "error_code",
			err: jerrors.New("test",
				source("testsource"),
				jerrors.WithCode("testcode"),
				WithCustomTrace("testservice", []string{"teststacktrace"}),
			),
		},
		{
			name: "context",
			ctx:  jlog.ContextWith(context.Background(), j.KS("ctx_key", "ctx_val")),
			err: jerrors.New("test",
				source("testsource"),
				WithCustomTrace("testservice", []string{"teststacktrace"}),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			jlog.SetDefaultLoggerForTesting(t, buf)
			jlog.Error(tc.ctx, tc.err, source("testsource"))

			verifyOutput(t, "error_"+tc.name, buf.Bytes())
		})
	}
}

func TestDeprecated(t *testing.T) {
	opts := []jlog.Option{source("testsource")}

	testCases := []struct {
		name   string
		format string
		vl     []interface{}
	}{
		{
			name:   "mixed_types",
			format: "%d, %s, %v",
			vl:     []interface{}{1, "2", false},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			buff := new(bytes.Buffer)
			bufln := new(bytes.Buffer)

			jlog.SetDefaultLoggerForTesting(t, buf, opts...)
			jlog.Print(tc.vl...)

			jlog.SetDefaultLoggerForTesting(t, buff, opts...)
			jlog.Printf(tc.format, tc.vl...)

			jlog.SetDefaultLoggerForTesting(t, bufln, opts...)
			jlog.Println(tc.vl...)

			verifyOutput(t, "print_"+tc.name, buf.Bytes())
			verifyOutput(t, "printf_"+tc.name, buff.Bytes())
			verifyOutput(t, "println_"+tc.name, bufln.Bytes())
		})
	}
}

func verifyOutput(t *testing.T, goldenFileName string, output []byte) {
	t.Helper()
	flag.Parse()
	goldenFilePath := path.Join("testdata", goldenFileName+".golden")

	if *writeGoldenFiles {
		err := os.WriteFile(goldenFilePath, output, 0o777)
		require.NoError(t, err)

		// Nothing to check if we're writing.
		return
	}

	contents, err := os.ReadFile(goldenFilePath)
	require.NoError(t, err, "Error reading golden file %s: %v", goldenFilePath, err)

	assert.Equal(t, string(contents), string(output))
}

func BenchmarkInfoCtx(b *testing.B) {
	var buf bytes.Buffer
	jlog.SetDefaultLoggerForTesting(b, &buf)

	ctx := context.Background()
	ctx = jlog.ContextWith(ctx, j.KV("key1", "v1"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		jlog.Info(ctx, "test message", j.KV("mykey", 123))
	}
}

func BenchmarkErrorCtx(b *testing.B) {
	var buf bytes.Buffer
	jlog.SetDefaultLoggerForTesting(b, &buf)

	ctx := context.Background()
	ctx = jlog.ContextWith(ctx, j.KV("key1", "v1"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := jerrors.New("my error", j.KV("key2", "v2"))
		jlog.Error(ctx, err, j.KV("mykey", 123))
	}
}

func BenchmarkStdLibLog(b *testing.B) {
	var buf bytes.Buffer
	l := stdlib_log.New(&buf, "", stdlib_log.LstdFlags)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Printf("hello k=%d", 123)
	}
}
