package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStack tests the stack trace including line numbers.
// Adding anything to this file might break the test.
func TestStack(t *testing.T) {
	err := stack(5)
	je, ok := err.(*JettisonError)
	require.True(t, ok)

	require.Equal(t, expected, je.Hops[0].StackTrace)
}

func stack(i int) error {
	if i == 0 {
		return New("stack")
	}
	return stack(i - 1)
}

var expected = []string{
	"github.com/luno/jettison/errors/stack_test.go:21",
	"github.com/luno/jettison/errors/stack_test.go:23",
	"github.com/luno/jettison/errors/stack_test.go:23",
	"github.com/luno/jettison/errors/stack_test.go:23",
	"github.com/luno/jettison/errors/stack_test.go:23",
	"github.com/luno/jettison/errors/stack_test.go:23",
	"github.com/luno/jettison/errors/stack_test.go:12",
	"testing/testing.go:909",
	"runtime/asm_amd64.s:1357",
}
