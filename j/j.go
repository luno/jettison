// Package j provides abridged aliases for jettison options.
package j

import (
	"fmt"

	"github.com/luno/jettison"
	"github.com/luno/jettison/errors"
)

// KV returns a jettison key value option the with default format of value.
func KV(key string, value interface{}) jettison.Option {
	return jettison.WithKeyValueString(key, fmt.Sprint(value))
}

// KS returns a jettison key value string option.
func KS(key string, value string) jettison.Option {
	return jettison.WithKeyValueString(key, value)
}

// MKV is a multi jettison key value option with default formats of values.
//
//  Usage:
//    log.InfoCtx(ctx, "msg", j.MKV{"k1": 1, "k2": "v"})
type MKV map[string]interface{}

func (m MKV) Apply(details jettison.Details) {
	for key, value := range m {
		details.SetKey(key, fmt.Sprint(value))
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

// C is an alias for jettison/errors.WithCode.
func C(code string) jettison.Option {
	return errors.WithCode(code)
}
