// Package j provides abridged aliases for jettison options with some
// best practices built in.
package j

import (
	"fmt"
	"reflect"

	"github.com/luno/jettison"
	"github.com/luno/jettison/errors"
)

// KV returns a jettison key value option the with default format of
// a simple value or fmt.Stringer implementation. Complex values
// like slices, maps, structs are not printed since it is considered
// bad practice.
func KV(key string, value interface{}) jettison.Option {
	return jettison.WithKeyValueString(key, sprint(value))
}

// KS returns a jettison key value string option.
func KS(key string, value string) jettison.Option {
	return jettison.WithKeyValueString(key, value)
}

// MKV is a multi jettison key value option with default formats of
// simple values or fmt.Stringer implementations. Complex values
// like slices, maps, structs are not printed since it is considered
// bad practice.
//
//  Usage:
//    log.InfoCtx(ctx, "msg", j.MKV{"k1": 1, "k2": "v"})
type MKV map[string]interface{}

func (m MKV) Apply(details jettison.Details) {
	for key, value := range m {
		details.SetKey(key, sprint(value))
	}
}

// MKS is a multi jettison key value string option.
//
//  Usage:
//    log.InfoCtx(ctx, "msg", j.MKS{"k1": "v1", "k2": "v2"})
type MKS map[string]string

func (m MKS) Apply(details jettison.Details) {
	for key, value := range m {
		details.SetKey(key, value)
	}
}

// C is an alias for jettison/errors.WithCode. Since this
// should only be used with sentinel errors it also clears the useless
// init-time stack trace allowing wrapping to add proper stack trace.
func C(code string) jettison.OptionFunc {
	return func(details jettison.Details) {
		errors.WithCode(code)(details)
		errors.WithoutStackTrace()(details)
	}
}

var nosprints = map[reflect.Kind]bool{
	reflect.Struct:        true,
	reflect.Map:           true,
	reflect.Slice:         true,
	reflect.Array:         true,
	reflect.Ptr:           true,
	reflect.UnsafePointer: true,
	reflect.Uintptr:       true,
	reflect.Func:          true,
	reflect.Chan:          true,
	reflect.Interface:     true,
}

func sprint(i interface{}) string {
	if i == nil {
		return "<nil>"
	}

	// Shortcut some simple types
	switch i.(type) {
	case bool:
		return fmt.Sprint(i)
	case int:
		return fmt.Sprint(i)
	case int64:
		return fmt.Sprint(i)
	case string:
		return fmt.Sprint(i)
	case fmt.Stringer:
		return fmt.Sprint(i)
	case fmt.Formatter:
		return fmt.Sprint(i)
	}
	k := reflect.TypeOf(i).Kind()
	if nosprints[k] {
		return "<" + k.String() + ">"
	}
	return fmt.Sprint(i)
}
