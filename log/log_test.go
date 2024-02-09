package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	stdlib_log "log"
	"testing"

	"github.com/go-stack/stack"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"

	jerrors "github.com/luno/jettison/errors"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
	"github.com/luno/jettison/trace"
)

//go:generate go test -update

type source string

func (s source) ApplyToLog(e *Entry) {
	e.Source = string(s)
}

func (s source) ApplyToError(je *internal.Error) {
	je.Source = string(s)
}

type logKV models.KeyValue

func (k logKV) ApplyToError(je *internal.Error) {
	je.KV = append(je.KV, models.KeyValue(k))
}

func (k logKV) ApplyToLog(entry *Entry) {
	entry.Parameters = append(entry.Parameters, models.KeyValue(k))
}

func (k logKV) ContextKeys() []models.KeyValue {
	return []models.KeyValue{models.KeyValue(k)}
}

func kv(key string, value any) logKV {
	return logKV{Key: key, Value: fmt.Sprint(value)}
}

// WithCustomTrace sets the stack trace of the current hop to the given value.
func WithCustomTrace(bin string, stack []string) jerrors.ErrorOption {
	return func(je *internal.Error) {
		je.Binary = bin
		je.StackTrace = stack
	}
}

func TestLog(t *testing.T) {
	testCases := []struct {
		name string
		msg  string
		ctx  context.Context
		opts []Option
	}{
		{
			name: "message_only",
			msg:  "test_message",
		},
		{
			name: "message_with_kv",
			msg:  "test_message",
			opts: []Option{
				kv("key", "value"),
			},
		},
		{
			name: "message_with_error_level",
			msg:  "test_message",
			opts: []Option{
				WithLevel(LevelError),
			},
		},
		{
			name: "message_with_unordered_parameters",
			msg:  "test_message",
			opts: []Option{
				kv("a", "c"),
				kv("c", "d"),
				kv("d", "c"),
				kv("c", "a"),
			},
		},
		{
			name: "message_with_error",
			msg:  "test_message",
			opts: []Option{
				WithError(jerrors.New("test",
					source("testsource"),
					WithCustomTrace("testservice", []string{"teststacktrace"}),
				)),
			},
		},
		{
			name: "message_with_context",
			ctx:  ContextWith(context.Background(), kv("ctx_key", "ctx_val")),
			msg:  "test_message",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			SetDefaultLoggerForTesting(t, buf, source("testsource"))
			Info(tc.ctx, tc.msg, tc.opts...)

			goldie.New(t).Assert(t, "log_"+tc.name, buf.Bytes())
		})
	}
}

func TestError(t *testing.T) {
	jerrors.SetTraceConfigTesting(t, jerrors.TestingConfig)
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
			ctx:  ContextWith(context.Background(), kv("ctx_key", "ctx_val")),
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
			SetDefaultLoggerForTesting(t, buf)
			Error(tc.ctx, tc.err, source("testsource"))

			goldie.New(t).Assert(t, "error_"+tc.name, buf.Bytes())
		})
	}
}

func TestDeprecated(t *testing.T) {
	opts := []Option{source("testsource")}

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

			SetDefaultLoggerForTesting(t, buf, opts...)
			Print(tc.vl...)

			SetDefaultLoggerForTesting(t, buff, opts...)
			Printf(tc.format, tc.vl...)

			SetDefaultLoggerForTesting(t, bufln, opts...)
			Println(tc.vl...)

			g := goldie.New(t)
			g.Assert(t, "print_"+tc.name, buf.Bytes())
			g.Assert(t, "printf_"+tc.name, buff.Bytes())
			g.Assert(t, "println_"+tc.name, bufln.Bytes())
		})
	}
}

func BenchmarkInfoCtx(b *testing.B) {
	var buf bytes.Buffer
	SetDefaultLoggerForTesting(b, &buf)

	ctx := context.Background()
	ctx = ContextWith(ctx, kv("key1", "v1"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Info(ctx, "test message", kv("mykey", 123))
	}
}

func BenchmarkErrorCtx(b *testing.B) {
	var buf bytes.Buffer
	SetDefaultLoggerForTesting(b, &buf)

	ctx := context.Background()
	ctx = ContextWith(ctx, kv("key1", "v1"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := jerrors.New("my error")
		Error(ctx, err, kv("mykey", 123))
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
		expEntry Entry
	}{
		{
			name: "single jerr",
			err: jerrors.New("hello",
				WithCustomTrace(
					"api",
					[]string{"updateDatabase", "doRequest"},
				),
				source("source.go"),
			),
			expEntry: Entry{
				ErrorObject: &ErrorObject{
					Message: "hello",
					Source:  "source.go",
					Stack:   []string{"api"},
					StackTrace: MakeElastic([]string{
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
					source("source_file.go"),
				), "outer",
				WithCustomTrace(
					"api",
					[]string{"callService", "doHTTP"},
				),
			),
			expEntry: Entry{ErrorObject: &ErrorObject{
				Message: "outer: inner",
				Source:  "source_file.go",
				Stack:   []string{"api", "service"},
				StackTrace: MakeElastic([]string{
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
			expEntry: Entry{ErrorObject: &ErrorObject{
				Message: io.EOF.Error(),
			}},
		},
		{
			name: "kvs",
			err: jerrors.Wrap(
				jerrors.New("a",
					kv("inner_key", "inner_value"),
					jerrors.WithoutStackTrace(),
					source("inner"),
				),
				"",
				kv("outer_key", "outer_value"),
				jerrors.WithoutStackTrace(),
				source("outer - overwritten"),
			),
			expEntry: Entry{ErrorObject: &ErrorObject{
				Message: "a",
				Source:  "inner",
				Parameters: []models.KeyValue{
					{Key: "outer_key", Value: "outer_value"},
					{Key: "inner_key", Value: "inner_value"},
				},
			}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var a Entry
			WithError(tc.err).ApplyToLog(&a)
			// get rid of the other fields, tested separately
			a = Entry{ErrorObject: a.ErrorObject, ErrorObjects: a.ErrorObjects}
			assert.Equal(t, tc.expEntry, a)
		})
	}
}

func TestAddErrors(t *testing.T) {
	jerrors.SetTraceConfigTesting(t, trace.StackConfig{
		RemoveLambdas:   true,
		TrimRuntime:     true,
		FormatReference: func(call stack.Call) string { return fmt.Sprintf("%n", call) },
	})

	testCases := []struct {
		name     string
		err      error
		expEntry Entry
	}{
		{name: "nil is empty"},
		{
			name: "wrapped",
			err: jerrors.Wrap(
				jerrors.New("one", jerrors.WithoutStackTrace()),
				"", jerrors.WithoutStackTrace(),
			),
			expEntry: Entry{ErrorObject: &ErrorObject{Message: "one", Source: "TestAddErrors"}},
		},
		{
			name: "joined errors",
			err: jerrors.Join(
				jerrors.New("one", jerrors.WithoutStackTrace()),
				jerrors.New("two", jerrors.WithoutStackTrace()),
			),
			expEntry: Entry{ErrorObjects: []ErrorObject{
				{Message: "one", Source: "TestAddErrors"},
				{Message: "two", Source: "TestAddErrors"},
			}},
		},
		{
			name: "joins in joins",
			err: jerrors.Join(
				jerrors.New("one", jerrors.WithoutStackTrace()),
				jerrors.Join(
					jerrors.New("two", jerrors.WithoutStackTrace()),
					jerrors.New("three", jerrors.WithoutStackTrace()),
				),
			),
			expEntry: Entry{ErrorObjects: []ErrorObject{
				{Message: "one", Source: "TestAddErrors"},
				{Message: "two", Source: "TestAddErrors"},
				{Message: "three", Source: "TestAddErrors"},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var e Entry
			addErrors(&e, tc.err)
			assert.Equal(t, tc.expEntry, e)
		})
	}
}
