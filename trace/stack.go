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
	// PackagesHidden will remove any calls in the stack from these packages
	// If a call matches a package in PackagesShown and PackagesHidden, it will be shown
	PackagesHidden []string
	// TrimRuntime will remove entries from the Go runtime
	TrimRuntime bool

	// FormatStack is the format for lines in the stack trace
	// The default will print the source reference and the function name
	FormatStack func(stack.Call) string
	// FormatReference is the formatter used when creating source code references
	FormatReference func(stack.Call) string
}

func (c StackConfig) shouldKeepCall(call stack.Call) bool {
	if c.RemoveLambdas {
		fnName := fmt.Sprintf("%n", call)
		if strings.Contains(fnName, ".func") {
			return false
		}
	}
	if len(c.PackagesShown) == 0 && len(c.PackagesHidden) == 0 {
		return true
	}
	pkgName := fmt.Sprintf("%+k", call)
	for _, p := range c.PackagesShown {
		if strings.HasPrefix(pkgName, p) {
			return true
		}
	}
	for _, p := range c.PackagesHidden {
		if strings.HasPrefix(pkgName, p) {
			return false
		}
	}
	return len(c.PackagesShown) == 0
}

func (c StackConfig) formatStackLine(call stack.Call) string {
	if c.FormatStack != nil {
		return c.FormatStack(call)
	}
	return fmt.Sprintf("%+v %n", call, call)
}

func (c StackConfig) formatReference(ref stack.Call) string {
	if c.FormatReference != nil {
		return c.FormatReference(ref)
	}
	return fmt.Sprintf("%+v", ref)
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
		res = append(res, config.formatStackLine(c))
		if len(res) >= maxDepth {
			break
		}
	}
	return res
}

// GetSourceCodeRef returns the callers source code reference
func GetSourceCodeRef(skip int, config StackConfig) string {
	return config.formatReference(stack.Caller(skip + 1))
}
