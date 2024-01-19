package errors

import (
	"fmt"
	"testing"

	"github.com/luno/jettison/trace"
)

var traceConfig trace.StackConfig

func SetTraceConfig(config trace.StackConfig) {
	if !traceConfig.IsZero() {
		panic(fmt.Sprintln("config has already been set", traceConfig, config))
	}
	traceConfig = config
}

func SetTraceConfigTesting(t testing.TB, config trace.StackConfig) {
	old := traceConfig
	t.Cleanup(func() {
		traceConfig = old
	})
	traceConfig = config
}

// getTrace will get the current binary and a stacktrace
// skip will omit a certain number of stack calls before getTrace
func getTrace(skip int) (string, []string) {
	// Skip GetStackTrace and getTrace
	return trace.CurrentBinary(), trace.GetStackTrace(skip+2, traceConfig)
}
