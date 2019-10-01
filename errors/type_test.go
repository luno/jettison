package errors_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	go_2_errors "golang.org/x/exp/errors"
	gstatus "google.golang.org/grpc/status"
)

func TestToFromStatus(t *testing.T) {
	testCases := []struct {
		name   string
		errors []models.Error
	}{
		{
			name: "single error, single param",
			errors: []models.Error{
				{
					Message: "msg",
					Source:  "source",
					Parameters: []models.KeyValue{
						{Key: "key", Value: "value"},
					},
				},
			},
		},
		{
			name: "single error, many param",
			errors: []models.Error{
				{
					Message: "msg",
					Source:  "source",
					Parameters: []models.KeyValue{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
		},
		{
			name: "many error, many param",
			errors: []models.Error{
				{
					Message: "msg1",
					Source:  "source1",
					Parameters: []models.KeyValue{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
				{
					Message: "msg2",
					Source:  "source2",
					Parameters: []models.KeyValue{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			s := errors.JettisonError{
				Hops: []models.Hop{
					{
						Binary: "service",
						Errors: tc.errors,
					},
				},
			}

			// Simulate going over the wire.
			status := s.GRPCStatus()
			statusErr := status.Err()
			statusFromErr, ok := gstatus.FromError(statusErr)
			if !assert.True(t, ok) {
				return
			}

			s2, err := errors.FromStatus(statusFromErr)
			assert.NoError(t, err)
			assert.Equal(t, len(s.Hops), len(s2.Hops))
			assert.Equal(t, s.Hops[0].Binary, s2.Hops[0].Binary)
			assert.Equal(t, len(s.Hops[0].Errors), len(s2.Hops[0].Errors))
			for i := range s.Hops[0].Errors {
				e1 := s.Hops[0].Errors[i]
				e2 := s2.Hops[0].Errors[i]

				assert.Equal(t, e1.Message, e2.Message)
				assert.Equal(t, e1.Source, e2.Source)
				assert.Equal(t, len(e1.Parameters), len(e2.Parameters))
				for j := range e1.Parameters {
					assert.Equal(t, e1.Parameters[j].Key, e2.Parameters[j].Key)
					assert.Equal(t, e1.Parameters[j].Value, e2.Parameters[j].Value)
				}
			}
		})
	}
}

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
	id0 := errors.New("id0")
	id1 := errors.New("id1", errors.WithCode("code1"))
	id2 := errors.Wrap(id1, "id2")
	id3 := errors.Wrap(id2, "id3", errors.WithCode("code3"))

	testCases := []struct {
		name     string
		err      error
		expCodes []string
	}{
		{
			name:     "default code, no wrap",
			err:      id0,
			expCodes: []string{"id0"},
		},
		{
			name:     "custom code, no wrap",
			err:      id1,
			expCodes: []string{"code1"},
		},
		{
			name:     "wrapped once",
			err:      id2,
			expCodes: []string{"id2", "code1"},
		},
		{
			name:     "wrapped twice",
			err:      id3,
			expCodes: []string{"code3", "id2", "code1"},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			je, ok := tc.err.(*errors.JettisonError)
			require.True(t, ok)

			assert.Equal(t, tc.expCodes, errors.GetCodes(je))
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
	err2 := errors.New("jettison error").(*errors.JettisonError)

	je = errors.Wrap(err0, "wrap").(*errors.JettisonError)
	assert.True(t, je.As(&err0))
	assert.True(t, go_2_errors.As(je, &err0))
	assert.False(t, je.As(&err1))

	je = errors.Wrap(err1, "wrap").(*errors.JettisonError)
	assert.True(t, je.As(&err1))
	assert.True(t, go_2_errors.As(je, &err1))
	assert.False(t, je.As(&err0))

	je = errors.Wrap(err2, "wrap").(*errors.JettisonError)
	assert.True(t, je.As(&err2))
	assert.True(t, go_2_errors.As(je, &err2))
	assert.False(t, je.As(&err0))
	assert.False(t, je.As(&err1))
}

func TestGetKey(t *testing.T) {
	err := errors.New("test", j.KV("key", "value")).(*errors.JettisonError)

	v, ok := err.GetKey("key")
	assert.True(t, ok)
	assert.Equal(t, "value", v)

	v, ok = err.GetKey("nonexistent")
	assert.False(t, ok)
	assert.Zero(t, v)
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
