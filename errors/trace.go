package errors

import (
	"fmt"

	"github.com/luno/jettison/trace"
)

var traceConfig trace.StackConfig

func SetTraceConfig(config trace.StackConfig) {
	if !traceConfig.IsZero() {
		panic(fmt.Sprintln("config has already been set", traceConfig, config))
	}
	traceConfig = config
}

// getTrace will get the current binary and a stacktrace
// skip will omit a certain number of stack calls before getTrace
func getTrace(skip int) (string, []string) {
	// Skip GetStackTrace and getTrace
	return trace.CurrentBinary(), trace.GetStackTrace(skip+2, traceConfig)
}

func hasTrace(err error) bool {
	errs := []error{err}
	for len(errs) > 0 {
		e := errs[0]
		errs = errs[1:]
		if je, ok := e.(*JettisonError); ok && je.Binary != "" {
			return true
		}
		switch unw := e.(type) {
		case interface{ Unwrap() error }:
			if err := unw.Unwrap(); err != nil {
				errs = append(errs, err)
			}
		case interface{ Unwrap() []error }:
			errs = append(errs, unw.Unwrap()...)
		}
	}
	return false
}
