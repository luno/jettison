package main

import (
	"context"

	"github.com/luno/jettison/_example/serverclient"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

func main() {
	// Set up example servers and client.
	s1 := serverclient.NewServer()
	s2 := serverclient.NewServer()
	c1 := serverclient.NewClient(s1.GetURL())
	c2 := serverclient.NewClient(s2.GetURL())
	defer func() {
		s1.Stop()
		s2.Stop()
	}()

	// Point the servers at each other for hopping.
	s1.SetClient(c2)
	s2.SetClient(c1)

	// Create a context that includes some default key-value pairs.
	ctx := log.ContextWith(context.Background(),
		j.KV("ctx_key", "ctx_value"),
		j.MKV{"ctx_key2": "ctx_value2", "ctx_key3": "ctx_value3"})

	err := c1.Hop(ctx, 2)
	log.Error(nil, err)
}
