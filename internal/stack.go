package internal

import (
	"fmt"
	"strings"

	"github.com/go-stack/stack"
)

const maxStackTraceDepth = 64

// GetStackTrace returns a rendered stacktrace of the calling code, skipping
// `skip` frames in the stack (1 is the GetStackTrace frame itself).
func GetStackTrace(skip int) []string {
	var res []string

	for i, c := range stack.Trace()[skip:] {
		if i >= maxStackTraceDepth {
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
