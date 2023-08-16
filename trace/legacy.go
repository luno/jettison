package trace

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/go-stack/stack"
)

// StripTestStacks strips "testing/testing.go:XXX" and "runtime/asm_XXXX.s:XXX"
// from stacktraces to avoid flappers on go version changes.
// Deprecated: use StackConfig.TrimRuntime
func StripTestStacks(_ *testing.T, b []byte) []byte {
	b = regexp.MustCompile("testing/testing\\.go:\\d+").ReplaceAll(b, []byte("testing/testing.go:X"))
	return regexp.MustCompile("runtime/asm_\\w+\\.s:\\d+").ReplaceAll(b, []byte("runtime/asm_X.s:X"))
}

// GetStackTraceLegacy returns a rendered stacktrace of the calling code, skipping
// `skip` frames in the stack (1 is the GetStackTrace frame itself).
// Deprecated: use GetStackTrace
func GetStackTraceLegacy(skip int) []string {
	var res []string

	for i, c := range stack.Trace()[skip:] {
		if i >= maxDepth {
			break
		}

		var fnName string
		if c.Frame().Func != nil {
			fnName = " " + sanitiseFnName(c.Frame().Func.Name())
		}

		res = append(res, fmt.Sprintf("%+v%s", c, fnName))
	}

	return res
}

// sanitiseFnName returns a function name including receiver (if applicable) but
// excluding package name.
func sanitiseFnName(name string) string {
	if lastslash := strings.LastIndex(name, "/"); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := strings.Index(name, "."); period >= 0 {
		name = name[period+1:]
	}

	name = strings.Replace(name, "·", ".", -1)
	return name
}
