package trace

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-stack/stack"
)

// PackagePath returns the package path for a type
// this will panic for any nil-like object
// Typically you would use this with a struct, such as testing.T or http.Server
func PackagePath(a any) string {
	return reflect.TypeOf(a).PkgPath()
}

type StackConfig struct {
	// RemoveLambdas will remove anonymous functions from the call stack
	RemoveLambdas bool
	// PackagesShown, if not empty, will limit the call stack to functions from these packages
	PackagesShown []string
	// TrimRuntime will remove entries from the Go runtime
	TrimRuntime bool
	// Format is the format for the stack.Call lines
	// The default will print the source reference and the function name
	Format func(stack.Call) string
}

func (c StackConfig) shouldKeepCall(call stack.Call) bool {
	if c.RemoveLambdas {
		fnName := fmt.Sprintf("%n", call)
		if strings.Contains(fnName, ".func") {
			return false
		}
	}
	if len(c.PackagesShown) == 0 {
		return true
	}
	pkgName := fmt.Sprintf("%+k", call)
	for _, p := range c.PackagesShown {
		if strings.HasPrefix(pkgName, p) {
			return true
		}
	}
	return false
}

func (c StackConfig) formatLine(call stack.Call) string {
	if c.Format != nil {
		return c.Format(call)
	}
	return fmt.Sprintf("%+v %n", call, call)
}

const maxDepth = 64

// GetStackTrace returns a rendered stacktrace of the calling code, skipping
// `skip` frames in the stack prior to this function
func GetStackTrace(skip int, config StackConfig) []string {
	var res []string
	trace := stack.Trace()
	if config.TrimRuntime {
		trace = trace.TrimRuntime()
	}
	for _, c := range trace[skip+1:] {
		if !config.shouldKeepCall(c) {
			continue
		}
		res = append(res, config.formatLine(c))
		if len(res) >= maxDepth {
			break
		}
	}
	return res
}

// GetSourceCodeRef returns the callers source code reference
func GetSourceCodeRef(skip int, config StackConfig) string {
	return config.Format(stack.Caller(skip + 1))
}
