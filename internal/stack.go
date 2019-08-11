package internal

import (
	"fmt"

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
		res = append(res, fmt.Sprintf("%+v", c))
	}

	return res
}
