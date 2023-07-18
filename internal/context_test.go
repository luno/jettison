package internal_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/j"
)

func TestContextWith(t *testing.T) {
	ctx := context.Background()

	assert.Empty(t, internal.ContextKeyValues(ctx))

	ctx = internal.ContextWith(ctx, j.KV("key", "value"))
	kvs := internal.ContextKeyValues(ctx)
	assert.Len(t, kvs, 1)

	// Check that the option was persisted correctly.
	assert.Equal(t, kvs[0].Key, "key")
	assert.Equal(t, kvs[0].Value, "value")

	ctx = internal.ContextWith(ctx, j.KV("key2", "value2"))
	kv2 := internal.ContextKeyValues(ctx)
	assert.Len(t, kv2, 2)

	// Check that both options were persisted.
	assert.Equal(t, kv2[0].Key, "key")
	assert.Equal(t, kv2[0].Value, "value")
	assert.Equal(t, kv2[1].Key, "key2")
	assert.Equal(t, kv2[1].Value, "value2")
}
