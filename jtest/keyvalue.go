package jtest

import (
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

// AssertKeyValues asserts that err contains at least the key values in kvs
//
// jtest.AssertKeyValues(t, j.MKS{"key":"value"}, err)
//
// If err contains keys that aren't in kvs they are ignored
// If err doesn't contain a key in kvs, then the test will fail
func AssertKeyValues(t testing.TB, expKVs j.MKS, err error, msg ...any) bool {
	t.Helper()

	var failure bool
	errKVs := errors.GetKeyValues(err)

	for expKey, expVal := range expKVs {
		val, has := errKVs[expKey]
		if !has {
			t.Error(failJKeyNotPresent(expKey, err, msg))
		} else if val != expVal {
			t.Error(failJKeyValuesMismatch(expKey, expVal, val, msg))
		} else {
			continue
		}
		failure = true
	}
	return !failure
}

// RequireKeyValues asserts that err contains at least the key values in kvs
//
// jtest.RequireKeyValues(t, j.MKS{"key":"value"}, err)
//
// If the assertion fails then the test will terminate immediately.
// See AssertKeyValues for more details
func RequireKeyValues(t testing.TB, kvs j.MKS, err error, msg ...any) bool {
	t.Helper()

	pass := AssertKeyValues(t, kvs, err, msg...)
	if !pass {
		t.FailNow()
	}
	return pass
}
