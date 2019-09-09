package log_test

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	stdlib_log "log"
	"os"
	"path"
	"testing"

	"github.com/luno/jettison"
	"github.com/luno/jettison/errors"
	jerrors "github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	jlog "github.com/luno/jettison/log"
	"github.com/stretchr/testify/assert"
)

var writeGoldenFiles = flag.Bool("write-golden-files", false,
	"Whether or not to overwrite golden files with test output.")

func TestLog(t *testing.T) {
	defer jlog.SetDefaultLoggerForTesting(t, os.Stdout)
	testCases := []struct {
		name string
		msg  string
		opts []jettison.Option
	}{
		{
			name: "message_only",
			msg:  "test_message",
		},
		{
			name: "message_with_kv",
			msg:  "test_message",
			opts: []jettison.Option{
				jettison.WithKeyValueString("key", "value"),
			},
		},
		{
			name: "message_with_error_level",
			msg:  "test_message",
			opts: []jettison.Option{
				jlog.WithLevel(jlog.LevelError),
			},
		},
		{
			name: "message_with_unordered_parameters",
			msg:  "test_message",
			opts: []jettison.Option{
				jettison.WithKeyValueString("a", "c"),
				jettison.WithKeyValueString("c", "d"),
				jettison.WithKeyValueString("d", "c"),
				jettison.WithKeyValueString("c", "a"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			jlog.SetDefaultLoggerForTesting(t, buf, jettison.WithSource("testsource"))
			jlog.Info(nil, tc.msg, tc.opts...)

			verifyOutput(t, "log_"+tc.name, buf.Bytes())
		})
	}
}

func TestError(t *testing.T) {
	defer jlog.SetDefaultLoggerForTesting(t, os.Stdout)
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "message_only",
			err: jerrors.New("test",
				jettison.WithSource("testsource"),
				jerrors.WithBinary("testservice"),
				errors.WithStackTrace([]string{"teststacktrace"})),
		},
		{
			name: "error_code",
			err: jerrors.New("test",
				jettison.WithSource("testsource"),
				jerrors.WithBinary("testservice"),
				jerrors.WithCode("testcode"),
				errors.WithStackTrace([]string{"teststacktrace"})),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			jlog.SetDefaultLoggerForTesting(t, buf)
			jlog.Error(nil, tc.err, jettison.WithSource("testsource"))

			verifyOutput(t, "error_"+tc.name, buf.Bytes())
		})
	}
}

func TestDeprecated(t *testing.T) {
	opts := []jettison.Option{jettison.WithSource("testsource")}
	defer jlog.SetDefaultLoggerForTesting(t, os.Stdout, opts...)

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
	flag.Parse()
	goldenFilePath := path.Join("testdata", goldenFileName)

	if *writeGoldenFiles {
		ioutil.WriteFile(goldenFilePath, output, 0777)

		// Nothing to check if we're writing.
		return
	}

	contents, err := ioutil.ReadFile(goldenFilePath)
	if err != nil {
		t.Errorf("Error reading golden file %s: %v", goldenFilePath, err)
	}

	assert.Equal(t, string(contents), string(output))
}

func BenchmarkInfoCtx(b *testing.B) {
	var buf bytes.Buffer
	defer jlog.SetDefaultLoggerForTesting(nil, &buf)

	ctx := context.Background()
	ctx = jlog.ContextWith(ctx, j.KV("key1", "v1"))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		jlog.Info(ctx, "test message", j.KV("mykey", 123))
	}
}

func BenchmarkErrorCtx(b *testing.B) {
	var buf bytes.Buffer
	defer jlog.SetDefaultLoggerForTesting(nil, &buf)

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
