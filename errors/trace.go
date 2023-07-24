package errors

import (
	"fmt"

	"github.com/luno/jettison/models"
	"github.com/luno/jettison/trace"
)

var traceConfig trace.StackConfig

func SetTraceConfig(config trace.StackConfig) {
	if !traceConfig.IsZero() {
		panic(fmt.Sprintln("config has already been set", traceConfig, config))
	}
	traceConfig = config
}

func getTrace() models.Hop {
	return models.Hop{
		Binary: trace.CurrentBinary(),
		// Skip GetStackTrace, getTrace, and New/Wrap
		StackTrace: trace.GetStackTrace(3, traceConfig),
	}
}
