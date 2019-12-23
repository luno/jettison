// Package jtest provides simple test assertion functions for the jettison
// errors package.
//
// The style is similar to the assert and require packages of the
// github.com/stretchr/testify library.
package jtest

import (
	"fmt"
	"testing"

	"github.com/luno/jettison/errors"
	"gopkg.in/yaml.v2"
)

// Assert asserts that the specified error matches the expected one. The test
// will be marked failed if it does not.
//
//    jtest.Assert(t, ErrWhatIExpect, err)
func Assert(t *testing.T, expected, actual error, msgs ...interface{}) bool {
	t.Helper()

	if !errors.Is(actual, expected) {
		t.Error(failLog(expected, actual, msgs))
		return false
	}
	return true
}

// Require asserts that the specified error matches the expected one. The test
// will be marked failed if it does not. It also stops test execution when it
// fails.
//
//    jtest.Require(t, ErrWhatIExpect, err)
func Require(t *testing.T, expected, actual error, msg ...interface{}) {
	t.Helper()
	if !Assert(t, expected, actual, msg...) {
		t.FailNow()
	}
}

func failLog(expected, actual error, msgs ...interface{}) string {
	l := fmt.Sprintf("No error in chain matches expected:\n"+
		"expected: %+v\n"+
		"actual:   %+v\n", pretty(expected), pretty(actual))

	message := messageFromMsgs(msgs...)
	if message != "" {
		l += fmt.Sprintf("message:  %+v\n", message)
	}

	return l
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

	if len(msgs) == 1 {
		m := msgs[0]
		if msgAsStr, ok := m.(string); ok {
			return msgAsStr
		}

		return fmt.Sprintf("%v", m)
	}

	if len(msgs) > 1 {
		var msg string

		for _, m := range msgs {
			if msg != "" {
				msg += " "
			}

			if msgAsStr, ok := m.(string); ok {
				msg += msgAsStr
				continue
			}

			msg += fmt.Sprintf("%v", m)
		}

		return msg
	}
	return ""
}
