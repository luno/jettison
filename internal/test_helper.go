package internal

import (
	"regexp"
	"testing"
)

// StripTestStacks strips "testing/testing.go:XXX" and "runtime/asm_XXXX.s:XXX"
// from stacktraces to avoid flapers on go version changes.
func StripTestStacks(_ *testing.T, b []byte) []byte {
	b = regexp.MustCompile("testing/testing\\.go:\\d+").ReplaceAll(b, []byte("testing/testing.go:X"))
	return regexp.MustCompile("runtime/asm_\\w+\\.s:\\d+").ReplaceAll(b, []byte("runtime/asm_X.s:X"))
}
