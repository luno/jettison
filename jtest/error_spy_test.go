package jtest

import (
	"fmt"
	"testing"
)

// testErrorSpy is a testing.TB that captures errors.
type testErrorSpy struct {
	testing.TB

	failed   bool
	messages []string
}

func newTestErrorSpy(t testing.TB) *testErrorSpy {
	return &testErrorSpy{TB: t}
}

func (t *testErrorSpy) Error(args ...any) {
	t.messages = append(t.messages, fmt.Sprintln(args...))
	t.failed = true
}
