package jtest

import (
	"fmt"
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

func TestAssertKeyValue(t *testing.T) {
	testCases := []struct {
		name string

		err error
		kvs map[string]string

		expPass bool
	}{
		{name: "nil error doesn't contain empty key", expPass: true},
		{
			name:    "nil error doesn't contain any key",
			kvs:     j.MKS{"key": "value"},
			expPass: false,
		},
		{
			name:    "New error contains key",
			err:     errors.New("hello", j.KV("key1", "value1")),
			kvs:     j.MKS{"key1": "value1"},
			expPass: true,
		},
		{
			name:    "New error contains some other key",
			err:     errors.New("hello", j.KV("key1", "value1")),
			kvs:     j.MKS{"key2": "value2"},
			expPass: false,
		},
		{
			name:    "New error contains different value",
			err:     errors.New("hello", j.KV("key1", "value1")),
			kvs:     j.MKS{"key1": "value2"},
			expPass: false,
		},
		{
			name:    "Wrapped error contains value",
			err:     errors.Wrap(errors.New("hello"), "", j.KV("key1", "value1")),
			kvs:     j.MKS{"key1": "value1"},
			expPass: true,
		},
		{
			name: "joined error contains value",
			err: errors.Join(
				fmt.Errorf("other error"),
				errors.New("hello", j.KV("key1", "value1")),
			),
			kvs:     j.MKS{"key1": "value1"},
			expPass: true,
		},
		{
			name: "can match many kvs",
			err:  errors.New("hello", j.KV("key1", "value1"), j.KV("key2", "value2")),
			kvs: j.MKS{
				"key1": "value1",
				"key2": "value2",
			},
			expPass: true,
		},
		{
			name: "extra keys in error is fine",
			err:  errors.New("hello", j.KV("key1", "value1"), j.KV("key2", "value2")),
			kvs: j.MKS{
				"key1": "value1",
			},
			expPass: true,
		},
		{
			name: "extra keys in test is not fine",
			err:  errors.New("hello", j.KV("key1", "value1")),
			kvs: j.MKS{
				"key1": "value1",
				"key2": "value2",
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockT := new(testing.T)

			pass := AssertKeyValues(mockT, tc.kvs, tc.err)
			if pass != tc.expPass {
				t.Errorf("Expected test result %v, got %v", tc.expPass, pass)
			}
			if mockT.Failed() == tc.expPass {
				t.Errorf("Expected test failure to be %v, got %v", tc.expPass, mockT.Failed())
			}
		})
	}
}
