package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-stack/stack"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"

	"github.com/luno/jettison/trace"
)

//go:generate go test -update

func TestSetTraceConfig(t *testing.T) {
	cfg := trace.StackConfig{
		RemoveLambdas: true,
		PackagesShown: []string{trace.PackagePath(JettisonError{})},
		TrimRuntime:   true,
		Format: func(call stack.Call) string {
			return fmt.Sprintf("%+k:%n", call, call)
		},
	}
	SetTraceConfig(cfg)
	_, st := getTrace(0)
	assert.Equal(t, []string{"github.com/luno/jettison/errors:TestSetTraceConfig"}, st)

	assert.Panics(t, func() {
		SetTraceConfig(trace.StackConfig{})
	})
}

// TestStack tests the stack trace including line numbers.
// Adding anything to this file might break the test.
func TestStack(t *testing.T) {
	SetTraceConfigTesting(t, trace.StackConfig{
		TrimRuntime: true,
		Format: func(call stack.Call) string {
			return fmt.Sprintf("%s %n", call, call)
		},
	})
	err := stackCalls(5)
	tr := []byte(strings.Join(err.StackTrace, "\n") + "\n")
	goldie.New(t).Assert(t, t.Name(), tr)
}

func stackCalls(i int) *JettisonError {
	if i == 0 {
		return New("stack").(*JettisonError)
	}
	return stackCalls(i - 1)
}
