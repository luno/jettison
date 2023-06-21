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
	jExpErr, expIsJError := expected.(*errors.JettisonError)
	if !expIsJError {
		// If expected error is not a jettison error, no need to compare further.
		return
	}

	expectedKeys := jExpErr.GetKeyValues()
	if len(expectedKeys) == 0 {
		// If we have no keys in the expected error, then we don't want to compare further.
		return
	}

	jActErr, actIsJError := actual.(*errors.JettisonError)
	if !actIsJError {
		// If actual error is not a jettison error, no need to compare further.
		return
	}

	// If both errs are jettison errors, we want to assert all the expected jettison keys are present in the actual error.
	for expKey, expValue := range expectedKeys {
		actualValue, present := jActErr.GetKey(expKey)
		if !present {
			t.Error(failJKeyNotPresent(expKey, actual, msgs))
			continue
		}

		if expValue != actualValue {
			t.Error(failJKeyValuesMismatch(expKey, expValue, actualValue, msgs))
		}
	}

	return
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
	l := fmt.Sprintf("No error in chain matches expected:\n"+
		"expected: %+v\n"+
		"actual:   %+v\n", pretty(expected), pretty(actual))

	return l + messageFromMsgs(msgs...)
}

func failNilLog(actual error, msgs ...interface{}) string {
	l := fmt.Sprintf("Unexpected non-nil error:\n"+
		"actual:   %+v\n", pretty(actual))

	return l + messageFromMsgs(msgs...)
}

func failJKeyNotPresent(key string, actual error, msgs ...interface{}) string {
	l := fmt.Sprintf("Expected jettison key '%v' was not present in actual error:\n"+
		"%+v\n", key, pretty(actual))

	return l + messageFromMsgs(msgs...)
}

func failJKeyValuesMismatch(key, expected, actual string, msgs ...interface{}) string {
	l := fmt.Sprintf("jettison values differ for key '%v':\n"+
		"expected: %+v\n"+
		"actual:   %+v\n", key, expected, actual)

	return l + messageFromMsgs(msgs...)
}

func pretty(err error) string {
	if err == nil {
		return fmt.Sprint(err)
	}

	msg := fmt.Sprintf("%+v", err)

	jerr := new(errors.JettisonError)
	if !errors.As(err, &jerr) {
		return msg
	}

	var val interface{}
	val = jerr
	if len(jerr.Hops) == 1 {
		val = jerr.Hops[0].Errors
	}

	b, err := yaml.Marshal(val)
	if err != nil {
		panic(err)
	}
	return msg + "\n" + string(b)
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
