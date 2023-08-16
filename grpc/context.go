package grpc

import (
	"context"
	"sort"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/luno/jettison/log"
	"github.com/luno/jettison/models"
)

var grpcPrefix = "__jettison__"

func fromJettisonKey(key string) (string, bool) {
	return strings.CutPrefix(key, grpcPrefix)
}

func toJettisonKey(key string) string {
	return grpcPrefix + key
}

func incomingContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	var kvs []models.KeyValue
	for k, vs := range md {
		key, ok := fromJettisonKey(k)
		if !ok {
			continue
		}
		for _, v := range vs {
			kvs = append(kvs, models.KeyValue{Key: key, Value: v})
		}
	}
	sort.Slice(kvs, func(i, j int) bool {
		if kvs[i].Key == kvs[j].Key {
			return kvs[i].Value < kvs[j].Value
		}
		return kvs[i].Key < kvs[j].Key
	})
	return log.ContextWithKeyValues(ctx, kvs)
}

func outgoingContext(ctx context.Context) context.Context {
	kvs := log.ContextKeyValues(ctx)
	if len(kvs) == 0 {
		return ctx
	}
	args := make([]string, 0, len(kvs)*2)
	for _, kv := range kvs {
		args = append(args, toJettisonKey(kv.Key), kv.Value)
	}
	return metadata.AppendToOutgoingContext(ctx, args...)
}
