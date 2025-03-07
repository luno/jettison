package log

import (
	"context"
	"github.com/luno/jettison/models"
	"slices"
)

// contextKey is used to index jettison options in the given
type contextKey struct{}

var key = contextKey{}

// ContextOption allows us to use the same type as an option
// for ContextWith as well as other jettison interfaces.
type ContextOption interface {
	ContextKeys() []models.KeyValue
}

// ContextWith returns a new context with the given jettison options appended
// to its key/value store. When a context containing jettison options is
// passed to Info or Error, the options are automatically applied to
// the resulting log.
func ContextWith(ctx context.Context, opts ...ContextOption) context.Context {
	if len(opts) == 0 {
		return ctx
	}
	add := opts[0].ContextKeys()
	for i := 1; i < len(opts); i++ {
		add = append(add, opts[i].ContextKeys()...)
	}
	return ContextWithKeyValues(ctx, add)
}

func ContextWithKeyValues(ctx context.Context, add []models.KeyValue) context.Context {
	kvs := contextKV(ctx)
	add = slices.DeleteFunc(add, func(kv models.KeyValue) bool {
		return slices.Contains(kvs, kv)
	})
	if len(add) == 0 {
		return ctx
	}
	// we need to make a new slice for the child context
	nkv := make([]models.KeyValue, len(kvs)+len(add))
	copy(nkv, kvs)
	copy(nkv[len(kvs):], add)
	return context.WithValue(ctx, key, nkv)
}

// ContextKeyValues returns the list of jettison key values options contained in the given context.
func ContextKeyValues(ctx context.Context) []models.KeyValue {
	kvs := contextKV(ctx)
	if len(kvs) == 0 {
		return nil
	}
	ret := make([]models.KeyValue, len(kvs))
	copy(ret, kvs)
	return ret
}

func contextKV(ctx context.Context) []models.KeyValue {
	if ctx == nil {
		return nil
	}
	kvs, _ := ctx.Value(key).([]models.KeyValue)
	if len(kvs) == 0 {
		return nil
	}
	return ctx.Value(key).([]models.KeyValue)
}
