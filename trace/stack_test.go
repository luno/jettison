package trace

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/go-stack/stack"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
)

//go:generate go test -update

func TestStackTrace(t *testing.T) {
	testCases := []struct {
		name      string
		pkgConfig StackConfig
		skip      int
	}{
		{
			name:      "all except runtime please",
			pkgConfig: StackConfig{TrimRuntime: true},
		},
		{
			name:      "jettison only",
			pkgConfig: StackConfig{PackagesShown: []string{PackagePath(StackConfig{})}},
		},
		{
			name:      "no jettison",
			pkgConfig: StackConfig{TrimRuntime: true, PackagesHidden: []string{PackagePath(StackConfig{})}},
		},
		{
			name: "package in both, still shows",
			pkgConfig: StackConfig{
				TrimRuntime:    true,
				PackagesShown:  []string{PackagePath(StackConfig{})},
				PackagesHidden: []string{PackagePath(StackConfig{})},
			},
		},
		{
			name:      "no lambdas",
			pkgConfig: StackConfig{RemoveLambdas: true, TrimRuntime: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.pkgConfig.FormatStack = func(call stack.Call) string {
				return fmt.Sprintf("%+s %n", call, call)
			}
			tr := callingFunction(tc.skip, tc.pkgConfig)
			goldie.New(t).Assert(t, t.Name(), []byte(strings.Join(tr, "\n")))
		})
	}
}

// callingFunction gives us a real function in the stack trace
func callingFunction(skip int, config StackConfig) []string {
	return GetStackTrace(skip, config)
}

func TestPkgPath(t *testing.T) {
	testCases := []struct {
		name     string
		a        any
		expPath  string
		expPanic bool
	}{
		{name: "int", a: 1},
		{name: "nil", a: nil, expPanic: true},
		{name: "nil interface", a: testing.TB(nil), expPanic: true},
		{name: "testing", expPath: "testing", a: testing.T{}},
		{name: "http func", expPath: "net/http", a: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {})},
		{name: "http", expPath: "net/http", a: http.Server{}},
		{name: "non-base package", a: assert.Assertions{}, expPath: "github.com/stretchr/testify/assert"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := func() {
				assert.Equal(t, tc.expPath, PackagePath(tc.a))
			}
			if tc.expPanic {
				assert.Panics(t, f)
			} else {
				f()
			}
		})
	}
}
