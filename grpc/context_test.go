package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"github.com/luno/jettison/models"
)

func TestOutgoingContext(t *testing.T) {
	testCases := []struct {
		name  string
		ctx   context.Context
		expMD metadata.MD
	}{
		{name: "empty context", ctx: context.Background()},
		{
			name: "kv",
			ctx: log.ContextWith(
				context.Background(),
				j.KV("key1", "value1"),
			),
			expMD: metadata.MD{
				"__jettison__key1": []string{"value1"},
			},
		},
		{
			name: "unrelated metadata",
			ctx: metadata.NewOutgoingContext(
				context.Background(),
				metadata.Pairs("a", "b"),
			),
			expMD: metadata.MD{
				"a": []string{"b"},
			},
		},
		{
			name: "unrelated metadata is untouched",
			ctx: metadata.NewOutgoingContext(
				log.ContextWith(context.Background(), j.KV("a", "c")),
				metadata.Pairs("a", "b"),
			),
			expMD: metadata.MD{
				"a":             []string{"b"},
				"__jettison__a": []string{"c"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := outgoingContext(tc.ctx)
			md, _ := metadata.FromOutgoingContext(ctx)
			assert.Equal(t, tc.expMD, md)
		})
	}
}

func TestIncomingContext(t *testing.T) {
	testCases := []struct {
		name   string
		ctx    context.Context
		expKVs []models.KeyValue
	}{
		{name: "empty", ctx: context.Background()},
		{name: "ignore unrelated keys", ctx: metadata.NewIncomingContext(
			context.Background(),
			metadata.Pairs("a", "b", "c", "d"),
		)},
		{
			name: "parse jettison kv",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs("__jettison__hello", "world"),
			),
			expKVs: []models.KeyValue{
				{Key: "hello", Value: "world"},
			},
		},
		{
			name: "parse multiple kv",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(
					"__jettison__hello", "world",
					"__jettison__ping", "pong",
					"__jettison__hello", "bob",
					"__jettison__empty", "",
				),
			),
			expKVs: []models.KeyValue{
				{Key: "empty", Value: ""},
				{Key: "hello", Value: "bob"},
				{Key: "hello", Value: "world"},
				{Key: "ping", Value: "pong"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := incomingContext(tc.ctx)
			kvs := log.ContextKeyValues(ctx)
			assert.Equal(t, tc.expKVs, kvs)
		})
	}
}
