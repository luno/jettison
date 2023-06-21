package jtest

import (
	goerrors "errors"
	"strings"
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

func Test_assertJettisonErrors(t *testing.T) {
	tests := []struct {
		name               string
		expected           error
		actual             error
		assertionSucceeds  bool
		expAssertionString string
	}{
		{
			name:              "Non-jettison expected error",
			expected:          goerrors.New("stdlib error"),
			assertionSucceeds: true,
		},
		{
			name:              "Non-jettison actual",
			expected:          errors.New("jettison error"),
			actual:            goerrors.New("stdlib error"),
			assertionSucceeds: true,
		},
		{
			name:              "Jettison errors equal (no keys)",
			expected:          errors.New("test error"),
			actual:            errors.New("test error"),
			assertionSucceeds: true,
		},
		{
			name:               "Expected error has j.KV key, actual doesn't - fails",
			expected:           errors.New("test error", j.KV("key1", "value")),
			actual:             errors.New("test error"),
			assertionSucceeds:  false,
			expAssertionString: "Expected jettison key 'key1' was not present in actual error:",
		},
		{
			name:               "Expected error has j.KV key, actual has different one",
			expected:           errors.New("test error", j.KV("key1", "value")),
			actual:             errors.New("test error", j.KV("key2", "value")),
			assertionSucceeds:  false,
			expAssertionString: "Expected jettison key 'key1' was not present in actual error:",
		},
		{
			name:               "Expected error has j.MKV key, actual doesn't - fails",
			expected:           errors.New("test error", j.MKV{"key1": "value1"}),
			actual:             errors.New("test error"),
			assertionSucceeds:  false,
			expAssertionString: "Expected jettison key 'key1' was not present in actual error:",
		},
		{
			name:               "Expected error has j.MKV keys, actual has one same and one different - fails",
			expected:           errors.New("test error", j.MKV{"key1": "value1", "key2": "value2"}),
			actual:             errors.New("test error", j.MKV{"key1": "value1", "key3": "value2"}),
			assertionSucceeds:  false,
			expAssertionString: "Expected jettison key 'key2' was not present in actual error:",
		},
		{
			name:              "Expected error doesn't have j.KV key, actual does - succeeds",
			expected:          errors.New("test error"),
			actual:            errors.New("test error", j.KV("key", "value")),
			assertionSucceeds: true,
		},
		{
			name:              "Expected error doesn't have j.MKV keys, actual does - succeeds",
			expected:          errors.New("test error"),
			actual:            errors.New("test error", j.MKV{"key1": "value1", "key2": "value2"}),
			assertionSucceeds: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTestErrorSpy(t)

			assertJettisonErrors(ts, tt.expected, tt.actual, nil)

			if ts.failed {
				for _, m := range ts.messages {
					if strings.Contains(m, tt.expAssertionString) {
						return
					}
				}
				t.Errorf("expected assertion string '%v' was not found in error messages: \n%v\n", tt.expAssertionString, ts.messages)
			} else if !tt.assertionSucceeds {
				t.Errorf("expected assertJettisonErrors to fail, but it succeeded")
			}
		})
	}
}
