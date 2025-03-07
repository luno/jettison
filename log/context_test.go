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
		{
			name: "dedupe kvs",
			ctx: log.ContextWith(context.Background(), j.MKV{
				"one": "1",
				"two": "2",
			}),
			opts: []log.ContextOption{j.MKV{
				"one":   "3!",
				"two":   "2",
				"three": "3",
			}},
			expKVs: []models.KeyValue{
				{Key: "one", Value: "1"},
				{Key: "two", Value: "2"},
				{Key: "one", Value: "3!"},
				{Key: "three", Value: "3"},
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

func TestChildDoesntChangeParent(t *testing.T) {
	parent := log.ContextWith(context.Background(), j.KV("one", "1"))
	child := log.ContextWith(parent, j.KV("two", "2"))

	expParent := []models.KeyValue{{Key: "one", Value: "1"}}
	assert.Equal(t, expParent, log.ContextKeyValues(parent))
	expChild := []models.KeyValue{
		{Key: "one", Value: "1"},
		{Key: "two", Value: "2"},
	}
	assert.Equal(t, expChild, log.ContextKeyValues(child))
}

func BenchmarkContextWith(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(1)
	ctx := log.ContextWith(context.Background(), j.KV("one", "1"))

	for range b.N {
		_ = log.ContextWith(ctx, j.KV("one", "1"))
	}
}
