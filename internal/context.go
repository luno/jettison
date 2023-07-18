package internal

import (
	"context"

	"github.com/luno/jettison/models"
)

// contextKey is used to index jettison options in the given
type contextKey struct{}

var key = contextKey{}

type ContextOption interface {
	ContextKeys() []models.KeyValue
}

// ContextWith returns a new context with the given jettison options added to
// its map of values. Note that only key-value options will be retained over
// the wire.
func ContextWith(ctx context.Context, opts ...ContextOption) context.Context {
	var add []models.KeyValue
	for _, o := range opts {
		add = append(add, o.ContextKeys()...)
	}
	return ContextWithKeyValues(ctx, add)
}

func ContextWithKeyValues(ctx context.Context, add []models.KeyValue) context.Context {
	if len(add) == 0 {
		return ctx
	}
	kvs := append(ContextKeyValues(ctx), add...)
	return context.WithValue(ctx, key, kvs)
}

// ContextKeyValues returns the list of jettison key values options contained in the given context.
func ContextKeyValues(ctx context.Context) []models.KeyValue {
	if ctx == nil {
		return nil
	}
	kvs, _ := ctx.Value(key).([]models.KeyValue)
	if len(kvs) == 0 {
		return nil
	}
	ret := make([]models.KeyValue, len(kvs))
	copy(ret, kvs)
	return ret
}
