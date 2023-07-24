package log

import (
	"context"

	"github.com/luno/jettison/models"
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
