package errors

import (
	"fmt"
	"testing"

	"github.com/go-stack/stack"

	"github.com/luno/jettison/trace"
)

var (
	configSet   bool
	traceConfig trace.StackConfig
)

func SetTraceConfig(config trace.StackConfig) {
	if configSet {
		panic(fmt.Sprintln("config has already been set", traceConfig, config))
	}
	traceConfig = config
	configSet = true
}

func SetTraceConfigTesting(t testing.TB, config trace.StackConfig) {
	old := traceConfig
	t.Cleanup(func() {
		traceConfig = old
	})
	traceConfig = config
}

var TestingConfig = trace.StackConfig{
	TrimRuntime:   true,
	RemoveLambdas: true,
	FormatStack: func(call stack.Call) string {
		return fmt.Sprintf("%s %n", call, call)
	},
	FormatReference: func(call stack.Call) string {
		return fmt.Sprintf("%s %n", call, call)
	},
}

// getTrace will get the current binary and a stacktrace
// skip will omit a certain number of stack calls before getTrace
func getTrace(skip int) (string, []string) {
	// Skip GetStackTrace and getTrace
	return trace.CurrentBinary(), trace.GetStackTrace(skip+1, traceConfig)
}

// getSourceCode will get the current
// Skip getSourceCode
func getSourceCode(skip int) string {
	return trace.GetSourceCodeRef(skip+1, traceConfig)
}
