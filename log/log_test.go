package log_test

import (
	"bytes"
	"context"
	"io"
	stdlib_log "log"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"

	jerrors "github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	jlog "github.com/luno/jettison/log"
	"github.com/luno/jettison/models"
)

//go:generate go test -update

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

			goldie.New(t).Assert(t, "log_"+tc.name, buf.Bytes())
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

			goldie.New(t).Assert(t, "error_"+tc.name, buf.Bytes())
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

			g := goldie.New(t)
			g.Assert(t, "print_"+tc.name, buf.Bytes())
			g.Assert(t, "printf_"+tc.name, buff.Bytes())
			g.Assert(t, "println_"+tc.name, bufln.Bytes())
		})
	}
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

func TestAddError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expEntry jlog.Entry
	}{
		{
			name: "single jerr",
			err: jerrors.New("hello",
				WithCustomTrace(
					"api",
					[]string{"updateDatabase", "doRequest"},
				)),
			expEntry: jlog.Entry{
				ErrorObject: &jlog.ErrorObject{
					Message: "hello",
					Source:  "github.com/luno/jettison/log/log_test.go:244",
					Stack:   []string{"api"},
					StackTrace: jlog.MakeElastic([]string{
						"updateDatabase",
						"doRequest",
					}),
				},
			},
		},
		{
			name: "two stacks",
			err: jerrors.Wrap(
				jerrors.New("inner",
					WithCustomTrace(
						"service",
						[]string{"update", "handleRequest"},
					),
				), "outer",
				WithCustomTrace(
					"api",
					[]string{"callService", "doHTTP"},
				),
			),
			expEntry: jlog.Entry{ErrorObject: &jlog.ErrorObject{
				Message: "outer: inner",
				Source:  "github.com/luno/jettison/log/log_test.go:264",
				Stack:   []string{"api", "service"},
				StackTrace: jlog.MakeElastic([]string{
					"update",
					"handleRequest",
					"api -> service",
					"callService",
					"doHTTP",
				}),
			}},
		},
		{
			name: "standard error",
			err:  io.EOF,
			expEntry: jlog.Entry{ErrorObject: &jlog.ErrorObject{
				Message: io.EOF.Error(),
			}},
		},
		{
			name: "kvs",
			err: jerrors.Wrap(
				jerrors.New("a",
					j.KV("inner_key", "inner_value"),
					jerrors.WithoutStackTrace(),
				),
				"",
				j.KV("outer_key", "outer_value"),
				jerrors.WithoutStackTrace(),
			),
			expEntry: jlog.Entry{ErrorObject: &jlog.ErrorObject{
				Message: "a",
				Parameters: []models.KeyValue{
					{Key: "outer_key", Value: "outer_value"},
					{Key: "inner_key", Value: "inner_value"},
				},
			}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var a jlog.Entry
			jlog.WithError(tc.err).ApplyToLog(&a)
			// get rid of the other fields, tested separately
			a = jlog.Entry{ErrorObject: a.ErrorObject, ErrorObjects: a.ErrorObjects}
			assert.Equal(t, tc.expEntry, a)
		})
	}
}
