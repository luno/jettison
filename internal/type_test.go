package internal_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
)

func TestError(t *testing.T) {
	testCases := []struct {
		name  string
		err   error
		expIn []string
	}{
		{
			name:  "unwrapped error",
			err:   &internal.Error{Message: "base: error msg"},
			expIn: []string{"base: error msg"},
		},
		{
			name:  "once-wrapped error",
			err:   &internal.Error{Message: "inner", Err: &internal.Error{Message: "base: error msg"}},
			expIn: []string{"inner: base: error msg"},
		},
		{
			name:  "twice-wrapped error",
			err:   &internal.Error{Message: "outer", Err: &internal.Error{Message: "inner", Err: &internal.Error{Message: "base: error msg"}}},
			expIn: []string{"outer: inner: base: error msg"},
		},
		{
			name:  "wrapped error with key/value pair",
			err:   &internal.Error{Message: "key/value", Err: &internal.Error{Message: "base: error msg"}, KV: []models.KeyValue{{Key: "key", Value: "value"}}},
			expIn: []string{"key/value: base: error msg"},
		},
		{
			name: "wrapped error with two key/value pairs",
			err: &internal.Error{Message: "key/values", Err: &internal.Error{Message: "base: error msg"}, KV: []models.KeyValue{
				{Key: "key", Value: "value"},
				{Key: "key2", Value: "value2"},
			}},
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

func TestFormat(t *testing.T) {
	testCases := []struct {
		name       string
		err        *internal.Error
		expError   string
		expString  string
		expVerbose string
	}{
		{
			name: "wrapped",
			err: &internal.Error{
				Message: "wrap three", Err: &internal.Error{
					Message: "wrap two", Err: &internal.Error{
						Message: "wrap one", KV: []models.KeyValue{{Key: "w", Value: "w1"}},
						Err: &internal.Error{
							Message: "root error",
							KV: []models.KeyValue{
								{Key: "p1", Value: "v1"},
								{Key: "p2", Value: "v2"},
							},
						},
					},
				},
			},
			expError:   "wrap three: wrap two: wrap one: root error",
			expString:  "wrap three: wrap two: wrap one: root error",
			expVerbose: "wrap three: wrap two: wrap one(w=w1): root error(p1=v1, p2=v2)",
		},
		{
			name:       "sql error",
			err:        &internal.Error{Message: "wrap sql error", Err: sql.ErrNoRows, KV: []models.KeyValue{{Key: "w", Value: "w1"}}},
			expError:   "wrap sql error: sql: no rows in result set",
			expString:  "wrap sql error: sql: no rows in result set",
			expVerbose: "wrap sql error(w=w1): sql: no rows in result set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expError, tc.err.Error())
			assert.Equal(t, tc.expString, fmt.Sprintf("%v", tc.err))
			assert.Equal(t, tc.expVerbose, fmt.Sprintf("%#v", tc.err))
		})
	}
}
