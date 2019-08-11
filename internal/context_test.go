package internal_test

import (
	"context"
	"testing"

	"github.com/luno/jettison"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
	"github.com/stretchr/testify/assert"
)

func TestContextWith(t *testing.T) {
	ctx := context.Background()

	opts := internal.ContextOptions(ctx)
	assert.Empty(t, opts)

	ctx = internal.ContextWith(ctx, jettison.WithKeyValueString("key", "value"))
	opts = internal.ContextOptions(ctx)
	assert.Len(t, opts, 1)

	// Check that the option was persisted correctly.
	var l models.Log
	for _, o := range opts {
		o.Apply(&l)
	}
	assert.Len(t, l.Parameters, 1)
	assert.Equal(t, l.Parameters[0].Key, "key")
	assert.Equal(t, l.Parameters[0].Value, "value")

	ctx = internal.ContextWith(ctx, jettison.WithKeyValueString("key2", "value2"))
	opts = internal.ContextOptions(ctx)
	assert.Len(t, opts, 2)

	// Check that both options were persisted.
	l = models.Log{}
	for _, o := range opts {
		o.Apply(&l)
	}
	assert.Len(t, l.Parameters, 2)
	assert.Equal(t, l.Parameters[0].Key, "key")
	assert.Equal(t, l.Parameters[0].Value, "value")
	assert.Equal(t, l.Parameters[1].Key, "key2")
	assert.Equal(t, l.Parameters[1].Value, "value2")
}

func TestContextDetails(t *testing.T) {
	opts := []jettison.Option{
		jettison.WithKeyValueString("key1", "value1"),
		jettison.WithKeyValueString("key2", "value2.1"),
		jettison.WithKeyValueString("key2", "value2.2"),
	}

	cd := make(internal.ContextDetails)
	for _, o := range opts {
		o.Apply(&cd)
	}

	// Should have set the key values on the ContextDetails.
	assert.Equal(t, map[string][]string{
		internal.ToJettisonKey("key1"): {"value1"},
		internal.ToJettisonKey("key2"): {"value2.1", "value2.2"},
	}, map[string][]string(cd.ToGrpcMetadata()))

	// Parsing back should give us the same options.
	resOpts := internal.FromGrpcMetadata(cd.ToGrpcMetadata())
	l := models.Log{}
	for _, o := range resOpts {
		o.Apply(&l)
	}
	assert.Len(t, l.Parameters, 3)
	assert.ElementsMatch(t, []models.KeyValue{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2.1"},
		{Key: "key2", Value: "value2.2"},
	}, l.Parameters)
}
