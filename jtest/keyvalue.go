package jtest

import (
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

func AssertKeyValues(t testing.TB, err error, kvs j.MKS, msg ...any) bool {
	t.Helper()

	var failure bool
	errKVs := errors.GetKeyValues(err)

	for expKey, expVal := range kvs {
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

func RequireKeyValues(t testing.TB, err error, kvs j.MKS, msg ...any) bool {
	t.Helper()

	pass := AssertKeyValues(t, err, kvs, msg...)
	if !pass {
		t.FailNow()
	}
	return pass
}
