// Package jtest provides simple test assertion functions for the jettison
// errors package.
//
// The style is similar to the assert and require packages of the
// github.com/stretchr/testify library.
package jtest

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
)

// Assert asserts that the specified error matches the expected one. The test
// will be marked failed if it does not.
//
//	jtest.Assert(t, ErrWhatIExpect, err)
func Assert(t testing.TB, expected, actual error, msgs ...interface{}) bool {
	t.Helper()

	if !errors.Is(actual, expected) {
		t.Error(failLog(expected, actual, msgs))
		return false
	}

	assertJettisonErrors(t, expected, actual, msgs...)
	if t.Failed() {
		return false
	}

	return true
}

// Require asserts that the specified error matches the expected one. The test
// will be marked failed if it does not. It also stops test execution when it
// fails.
//
//	jtest.Require(t, ErrWhatIExpect, err)
func Require(t testing.TB, expected, actual error, msg ...interface{}) {
	t.Helper()

	if !Assert(t, expected, actual, msg...) {
		t.FailNow()
	}
}

func assertJettisonErrors(t testing.TB, expected, actual error, msgs ...interface{}) {
	expectedKeys := errors.GetKeyValues(expected)
	if len(expectedKeys) == 0 {
		// If we have no keys in the expected error, then we don't want to compare further.
		return
	}
	// If actual error is not a jettison error, no need to compare further.
	if !errors.As(actual, new(*internal.Error)) {
		return
	}
	actualKeys := errors.GetKeyValues(actual)

	// If both errs are jettison errors, we want to assert all the expected jettison keys are present in the actual error.
	for expKey, expValue := range expectedKeys {
		actualValue, present := actualKeys[expKey]
		if !present {
			t.Error(failJKeyNotPresent(expKey, actual, msgs))
			continue
		}
		if expValue != actualValue {
			t.Error(failJKeyValuesMismatch(expKey, expValue, actualValue, msgs))
		}
	}
}

// AssertNil asserts that the specified error is nil. The test will be marked
// failed if it does not. It is shorthand for `jtest.Assert(t, nil, err)`,
// although it provides slightly clearer failure output.
//
//	jtest.AssertNil(t, err)
func AssertNil(t testing.TB, actual error, msgs ...interface{}) bool {
	t.Helper()

	if actual != nil {
		t.Error(failNilLog(actual, msgs))
		return false
	}
	return true
}

// RequireNil asserts that the specified error is nil. The test will be marked
// failed if it does not, and execution will be stopped. It is shorthand for
// `jtest.Require(t, nil, err)`, although it provides slightly clearer failure
// output.
//
//	jtest.RequireNil(t, err)
func RequireNil(t testing.TB, actual error, msg ...interface{}) {
	t.Helper()

	if !AssertNil(t, actual, msg...) {
		t.FailNow()
	}
}

func failLog(expected, actual error, msgs ...interface{}) string {
	l := fmt.Sprintf("No error in chain matches expected:\nexpected: %+vactual:   %+v", pretty(expected), pretty(actual))

	return l + messageFromMsgs(msgs...)
}

func failNilLog(actual error, msgs ...interface{}) string {
	l := fmt.Sprintf("Unexpected non-nil error:\n"+
		"actual:   %+v", pretty(actual))

	return l + messageFromMsgs(msgs...)
}

func failJKeyNotPresent(key string, actual error, msgs ...interface{}) string {
	l := fmt.Sprintf("Expected jettison key '%v' was not present in actual error:\n"+
		"%+v", key, pretty(actual))

	return l + messageFromMsgs(msgs...)
}

func failJKeyValuesMismatch(key, expected, actual string, msgs ...interface{}) string {
	l := fmt.Sprintf("jettison values differ for key '%v':\n"+
		"expected: %+v\n"+
		"actual:   %+v\n", key, expected, actual)

	return l + messageFromMsgs(msgs...)
}

type prettyError struct {
	Message string            `yaml:"message,omitempty"`
	Code    string            `yaml:"code,omitempty"`
	KV      []models.KeyValue `yaml:"kv,omitempty"`
	// TODO(adam): Add source
}

func pretty(err error) string {
	if err == nil {
		return fmt.Sprint(err)
	}

	paths := errors.Flatten(err)
	pretties := make([]prettyError, 0, len(paths))
	for _, p := range paths {
		var pret prettyError
		if len(paths) > 1 {
			pret.Message = p[len(p)-1].Error()
		} else {
			// Can use the fully wrapped message if the error isn't joined
			pret.Message = p[0].Error()
		}
		for _, e := range p {
			je, ok := e.(*internal.Error)
			if !ok {
				continue
			}
			// Take the lowest
			if je.Code != "" {
				pret.Code = je.Code
			}
			pret.KV = append(pret.KV, je.KV...)
		}
		pretties = append(pretties, pret)
	}

	b, err := yaml.Marshal(pretties)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func messageFromMsgs(msgs ...interface{}) string {
	if len(msgs) == 0 {
		return ""
	}

	msg := "message:  "

	if len(msgs) == 1 {
		m := msgs[0]

		return msg + fmt.Sprintf("%v\n", m)
	}

	for i, m := range msgs {
		if i > 0 {
			msg += " "
		}

		msg += fmt.Sprintf("%v", m)
	}

	return msg + "\n"
}
