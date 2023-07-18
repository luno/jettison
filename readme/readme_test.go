// Package readme contains example code that's referenced in jettison's
// README.md template. The template generation is done using the `rebecca`
// tool: https://github.com/dave/rebecca
package readme

import (
	"context"
	"fmt"

	"github.com/luno/jettison"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

func ExampleNew() {
	// Construct your error as usual, with additional metadata.
	err := errors.New("something went wrong",
		j.KV("key", "value"),
		jettison.WithSource("Example()"))

	// Wrap errors with additional metadata as they get passed down the stack.
	err = errors.Wrap(err, "something else went wrong",
		j.KV("another_key", "another_value"))

	// Pass it around - including over gRPC - like you would any other error.
}

func ExampleInfo() {
	ctx := context.Background()

	// You can log general info as you normally would.
	log.Info(ctx, "entering the example function",
		j.KV("key", "value"))
}

func ExampleError() {
	ctx := context.Background()

	err := errors.New("a jettison error",
		j.KV("key", "value"),
		jettison.WithSource("Example()"))

	// Errors can be logged separately, with metadata marshalled to JSON in
	// a machine-friendly manner.
	log.Error(ctx, err,
		j.KV("another_key", "another_value"))
}

func ExampleKS() {
	err := errors.New("using j.KS",
		j.KS("string_key", "value"))

	fmt.Printf("%%+v: %+v\n", err)
	fmt.Printf("%%#v: %#v\n", err)
	// Output:
	// %+v: using j.KS(string_key=value)
	// %#v: using j.KS(string_key=value)
}

func ExampleKV() {
	err := errors.New("using j.KV",
		j.KV("int_key", 1))

	fmt.Printf("%%+v: %+v\n", err)
	fmt.Printf("%%#v: %#v\n", err)
	// Output:
	// %+v: using j.KV(int_key=1)
	// %#v: using j.KV(int_key=1)
}
