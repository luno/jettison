package errors_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

func TestError(t *testing.T) {
	err1 := errors.New("base: error msg")
	err2 := errors.Wrap(err1, "inner")
	err3 := errors.Wrap(err2, "outer")
	err4 := errors.Wrap(err1, "key/value", j.KV("key", "value"))
	err5 := errors.Wrap(err1, "key/values", j.MKV{
		"key":  "value",
		"key2": "value2",
	})

	testCases := []struct {
		name  string
		err   error
		expIn []string
	}{
		{
			name:  "unwrapped error",
			err:   err1,
			expIn: []string{"base: error msg"},
		},
		{
			name:  "once-wrapped error",
			err:   err2,
			expIn: []string{"inner: base: error msg"},
		},
		{
			name:  "twice-wrapped error",
			err:   err3,
			expIn: []string{"outer: inner: base: error msg"},
		},
		{
			name:  "wrapped error with key/value pair",
			err:   err4,
			expIn: []string{"key/value: base: error msg"},
		},
		{
			name: "wrapped error with two key/value pairs",
			err:  err5,
			expIn: []string{
				"key/values: base: error msg",
				"key/values: base: error msg",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, tc.expIn, tc.err.Error())
		})
	}
}

func TestUnwrap(t *testing.T) {
	testCases := []struct {
		name     string
		err      *errors.JettisonError
		expCodes []string
	}{
		{
			name:     "default code, no wrap",
			err:      errors.New("id0").(*errors.JettisonError),
			expCodes: []string{"id0"},
		},
		{
			name:     "custom code, no wrap",
			err:      errors.New("id1", errors.WithCode("code1")).(*errors.JettisonError),
			expCodes: []string{"code1"},
		},
		{
			name: "wrapped once",
			err: errors.Wrap(
				errors.New("id1", errors.WithCode("code1")),
				"id2",
			).(*errors.JettisonError),
			expCodes: []string{"id2", "code1"},
		},
		{
			name: "wrapped twice",
			err: errors.Wrap(
				errors.Wrap(
					errors.New("id1", errors.WithCode("code1")),
					"id2",
				),
				"id3",
				errors.WithCode("code3"),
			).(*errors.JettisonError),
			expCodes: []string{"code3", "id2", "code1"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			codes := errors.GetCodes(tc.err)
			assert.Equal(t, tc.expCodes, codes)
		})
	}
}

type testErr string

func (te testErr) Error() string {
	return string(te)
}

type testErrPtr string

func (tep *testErrPtr) Error() string {
	return string(*tep)
}

func TestAs(t *testing.T) {
	tep := testErrPtr("custom error type with pointer")

	var je *errors.JettisonError
	err0 := testErr("custom error type")
	err1 := &tep
	err2 := errors.New("jettison error")

	je = errors.Wrap(err0, "wrap").(*errors.JettisonError)
	assert.True(t, errors.As(je, &err0))
	assert.False(t, errors.As(je, &err1))

	je = errors.Wrap(err1, "wrap").(*errors.JettisonError)
	assert.True(t, errors.As(je, &err1))
	assert.False(t, errors.As(je, &err0))

	je = errors.Wrap(err2, "wrap").(*errors.JettisonError)
	assert.True(t, errors.As(je, &err2))
	assert.False(t, errors.As(je, &err0))
	assert.False(t, errors.As(je, &err1))
}

func TestFormat(t *testing.T) {
	err1 := errors.New("root error", j.MKV{"p1": "v1", "p2": "v2"})
	err2 := errors.Wrap(err1, "wrap one", j.KV("w", "w1"))
	err3 := errors.Wrap(err2, "wrap two")
	err4 := errors.Wrap(err3, "wrap three")

	assert.Equal(t, "wrap three: wrap two: wrap one: root error", err4.Error())
	assert.Equal(t, "wrap three: wrap two: wrap one: root error", fmt.Sprintf("%v", err4))
	assert.Equal(t, "wrap three: wrap two: wrap one(w=w1): root error(p1=v1, p2=v2)", fmt.Sprintf("%+v", err4))

	err5 := errors.Wrap(sql.ErrNoRows, "wrap sql error", j.KV("w", "w1"))

	assert.Equal(t, "wrap sql error: sql: no rows in result set", err5.Error())
	assert.Equal(t, "wrap sql error: sql: no rows in result set", fmt.Sprintf("%s", err5))
	assert.Equal(t, "wrap sql error(w=w1): sql: no rows in result set", fmt.Sprintf("%#v", err5))
}
