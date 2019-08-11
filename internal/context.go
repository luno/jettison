package internal

import (
	"context"
	"strings"

	"github.com/luno/jettison"
	"google.golang.org/grpc/metadata"
)

// contextKey is used to index jettison options in the given
type contextKey struct{}

var key = contextKey{}

// ContextWith returns a new context with the given jettison options added to
// its map of values. Note that only key-value options will be retained over
// the wire.
func ContextWith(ctx context.Context, opts ...jettison.Option) context.Context {
	ctxOpts := ContextOptions(ctx)

	// NOTE(guy): The full slice below is to prevent a situation where two
	// child contexts end up overwriting each other's memory accidentally
	// because the slice has extra capacity. The idea is to cap the slice
	// capacity to the result length to force a reallocation next time.
	res := append(ctxOpts, opts...)
	return context.WithValue(ctx, key, res[:len(res):len(res)])
}

// ContextOptions returns the list of jettison options contained in the given
// context's map of values.
func ContextOptions(ctx context.Context) []jettison.Option {
	if ctx == nil {
		return nil
	}

	var opts []jettison.Option
	if optsI := ctx.Value(key); optsI != nil {
		optsV, ok := optsI.([]jettison.Option)
		if ok {
			opts = optsV
		}
	}

	return opts
}

// ContextDetails is an implementation of jettison.Details that encodes
// jettison key/value pairs in a format usable as gRPC metadata.
type ContextDetails metadata.MD

func (cd *ContextDetails) SetKey(key, value string) {
	key = ToJettisonKey(key)
	(*cd)[key] = append((*cd)[key], value)
}

func (cd *ContextDetails) SetSource(src string) {
	// not supported
}

func (cd *ContextDetails) ToGrpcMetadata() metadata.MD {
	return metadata.MD(*cd)
}

// FromGrpcMetadata parses out any jettison options listed in the given gRPC
// metadata.
func FromGrpcMetadata(md metadata.MD) []jettison.Option {
	var res []jettison.Option
	for k, vs := range md {
		if !isJettisonKey(k) {
			continue
		}

		key := FromJettisonKey(k)
		for _, v := range vs {
			res = append(res, jettison.WithKeyValueString(key, v))
		}
	}

	return res
}

var grpcPrefix = "__jettison__"

func FromJettisonKey(key string) string {
	if !isJettisonKey(key) {
		return key
	}

	return key[len(grpcPrefix):]
}

func ToJettisonKey(key string) string {
	return grpcPrefix + key
}

func isJettisonKey(key string) bool {
	return strings.HasPrefix(key, grpcPrefix)
}

var _ jettison.Details = new(ContextDetails)
