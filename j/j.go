// Package j provides abridged aliases for jettison options with some
// best practices built in.
package j

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/log"
	"github.com/luno/jettison/models"
)

var (
	allowedChars    = "0123456789abcdefghijklmnopqrstuvwxyz-_."
	allowedCharsMap map[rune]bool
)

func init() {
	allowedCharsMap = make(map[rune]bool)
	for _, ch := range allowedChars {
		allowedCharsMap[ch] = true
	}
}

// KV returns a jettison key value option the with default format of
// a simple value or fmt.Stringer implementation. Complex values
// like slices, maps, structs are not printed since it is considered
// bad practice.
func KV(key string, value any) MKV {
	return MKV{key: value}
}

// KS returns a jettison key value string option.
func KS(key string, value string) MKV {
	return MKV{key: value}
}

// MKV is a multi jettison key value option with default formats of
// simple values or fmt.Stringer implementations. Complex values
// like slices, maps, structs are not printed since it is considered
// bad practice.
//
//	Usage:
//	  log.InfoCtx(ctx, "msg", j.MKV{"k1": 1, "k2": "v"})
type MKV map[string]any

func (m MKV) ContextKeys() []models.KeyValue {
	res := make([]models.KeyValue, 0, len(m))
	for k, v := range m {
		res = append(res, models.KeyValue{Key: normalise(k), Value: sprint(v)})
	}
	return res
}

func (m MKV) ApplyToLog(l *log.Entry) {
	l.Parameters = append(l.Parameters, m.ContextKeys()...)
}

func (m MKV) ApplyToError(je *errors.JettisonError) {
	kvs := m.ContextKeys()
	je.KV = append(je.KV, kvs...)
	if len(je.Hops) == 0 {
		return
	}
	h := je.Hops[0]
	for _, kv := range kvs {
		h.SetKey(kv.Key, kv.Value)
	}
}

// MKS is a multi jettison key value string option.
//
//	Usage:
//	  log.InfoCtx(ctx, "msg", j.MKS{"k1": "v1", "k2": "v2"})
type MKS map[string]string

func (m MKS) ContextKeys() []models.KeyValue {
	res := make([]models.KeyValue, 0, len(m))
	for k, v := range m {
		res = append(res, models.KeyValue{Key: normalise(k), Value: v})
	}
	return res
}

func (m MKS) ApplyToLog(l *log.Entry) {
	l.Parameters = append(l.Parameters, m.ContextKeys()...)
}

func (m MKS) ApplyToError(je *errors.JettisonError) {
	kvs := m.ContextKeys()
	je.KV = append(je.KV, kvs...)
	if len(je.Hops) == 0 {
		return
	}
	h := je.Hops[0]
	for _, kv := range kvs {
		h.SetKey(kv.Key, kv.Value)
	}
}

// C is an alias for jettison/errors.WithCode. Since this
// should only be used with sentinel errors it also clears the useless
// init-time stack trace allowing wrapping to add proper stack trace.
func C(code string) errors.Option {
	return errors.C(code)
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

// normalise modifies the given key to conform to gRPC metadata requirements,
// as the keys have to be transmittable over the wire (in contexts, for
// instance).
// See https://godoc.org/google.golang.org/grpc/metadata#New.
func normalise(key string) string {
	// Uppercase characters are normalised to lower case.
	key = strings.ToLower(key)

	// Keys beginning with 'grpc-' are disallowed.
	key = strings.TrimPrefix(key, "grpc-")

	var res string
	for _, ch := range key {
		// Remove illegal characters from the key.
		if !allowedCharsMap[ch] {
			continue
		}

		res += string(ch)
	}

	return res
}
