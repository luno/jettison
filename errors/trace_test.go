package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-stack/stack"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luno/jettison/internal"
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
	err := stackCalls(5)
	je, ok := err.(*JettisonError)
	require.True(t, ok)

	tr := []byte(strings.Join(je.Hops[0].StackTrace, "\n") + "\n")
	tr = internal.StripTestStacks(t, tr)
	goldie.New(t).Assert(t, t.Name(), tr)
}

func stackCalls(i int) error {
	if i == 0 {
		return New("stack")
	}
	return stackCalls(i - 1)
}
