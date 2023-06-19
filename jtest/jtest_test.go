package jtest

import "testing"

func Test_assertJettisonErrors(t *testing.T) {
	tests := []struct {
		name     string
		expected error
		actual   error
		msg      []interface{}
	}{
		{
			name: "Non-jettison errors equal",
		},
		{
			name: "Non-jettison errors not equal",
		},
		{
			name: "Jettison errors equal",
		},
		{
			name: "Jettison errors not equal",
		},
		{
			name: "Expected error has keys actual doesn't",
		},
		{
			name: "Actual error has keys expected doesn't",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertJettisonErrors(t, tt.expected, tt.actual, nil)
		})
	}
}
