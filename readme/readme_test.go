// Package readme contains example code that's referenced in jettison's
// README.md template. The template generation is done using the `rebecca`
// tool: https://github.com/dave/rebecca
package readme

import (
	"context"

	"github.com/luno/jettison"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

func ExampleErrors() error {
	// Construct your error as usual, with additional metadata.
	err := errors.New("something went wrong",
		jettison.WithKeyValueString("key", "value"),
		jettison.WithSource("Example()"))

	// Wrap errors with additional metadata as they get passed down the stack.
	err = errors.Wrap(err, "something else went wrong",
		jettison.WithKeyValueString("another_key", "another_value"))

	// Pass it around - including over gRPC - like you would any other error.
	return err
}

func ExampleLog(ctx context.Context) {
	// You can log general info as you normally would.
	log.Info(ctx, "entering the example function",
		jettison.WithKeyValueString("key", "value"))

	err := errors.New("a jettison error",
		jettison.WithKeyValueString("key", "value"),
		jettison.WithSource("Example()"))

	// Errors can be logged separately, with metadata marshalled to JSON in
	// a machine-friendly manner.
	log.Error(ctx, err,
		jettison.WithKeyValueString("another_key", "another_value"))
}

func ExampleUtilities() error {
	return errors.New("using the aliases",
		j.KV("int_key", 1),
		j.KS("string_key", "value"))
}
