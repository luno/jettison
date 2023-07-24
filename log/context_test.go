package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"github.com/luno/jettison/models"
)

func TestContextWith(t *testing.T) {
	testCases := []struct {
		name   string
		ctx    context.Context
		opts   []log.ContextOption
		expKVs []models.KeyValue
	}{
		{name: "empty"},
		{
			name:   "single",
			ctx:    context.Background(),
			opts:   []log.ContextOption{j.KV("key", "value")},
			expKVs: []models.KeyValue{{Key: "key", Value: "value"}},
		},
		{
			name: "multiple retain order",
			ctx:  context.Background(),
			opts: []log.ContextOption{
				j.KV("key1", "value1"),
				j.KV("key2", "value"),
				j.KV("key1", "value2"),
			},
			expKVs: []models.KeyValue{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value"},
				{Key: "key1", Value: "value2"},
			},
		},
		{
			name: "appended kvs",
			ctx:  log.ContextWith(context.Background(), j.KV("one", "1")),
			opts: []log.ContextOption{j.MKV{"two": "2"}},
			expKVs: []models.KeyValue{
				{Key: "one", Value: "1"},
				{Key: "two", Value: "2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := log.ContextWith(tc.ctx, tc.opts...)
			assert.Equal(t, tc.expKVs, log.ContextKeyValues(ctx))
		})
	}
}
