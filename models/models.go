// Package models contains representations of Jettison objects that are passed
// to loggers.
package models

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/go-stack/stack"
)

type Hop struct {
	Binary     string   `json:"binary" protocp:"1"`
	StackTrace []string `json:"stack_trace,omitempty" protocp:"3"`
	Errors     []Error  `json:"errors,omitempty" protocp:"2"`
}

func (h *Hop) IsZero() bool {
	return h.Binary == "" && len(h.StackTrace) == 0 && len(h.Errors) == 0
}

// SetKey updates the parameters of the most recently added error in the hop.
func (h *Hop) SetKey(key, value string) {
	if h == nil || len(h.Errors) == 0 {
		return
	}

	h.Errors[0].Parameters = append(h.Errors[0].Parameters, KeyValue{
		Key:   key,
		Value: value,
	})

	sort.Slice(h.Errors[0].Parameters, func(i, j int) bool {
		return h.Errors[0].Parameters[i].Key < h.Errors[0].Parameters[j].Key
	})
}

// SetSource updates the source of the most recently added error in the hop.
func (h *Hop) SetSource(src string) {
	if h == nil || len(h.Errors) == 0 {
		return
	}

	h.Errors[0].Source = src
}

// Clone returns a copy of the original hop that can be mutated safely.
func (h *Hop) Clone() Hop {
	res := *h

	res.Errors = nil
	for _, e := range h.Errors {
		res.Errors = append(res.Errors, e.Clone())
	}

	return res
}

func NewHop() Hop {
	return Hop{Binary: filepath.Base(os.Args[0])}
}

type Error struct {
	Code       string     `json:"code,omitempty" protocp:"4"`
	Message    string     `json:"message" protocp:"1"`
	Source     string     `json:"source" protocp:"2"`
	Parameters []KeyValue `json:"parameters,omitempty" protocp:"3"`
}

// Clone returns a copy of the original error that can be mutated safely.
func (e *Error) Clone() Error {
	res := *e

	res.Parameters = make([]KeyValue, len(e.Parameters))
	copy(res.Parameters, e.Parameters)

	return res
}

func NewError(msg string) Error {
	return Error{
		Message: msg,
		Source:  fmt.Sprintf("%+v", stack.Caller(2)),
		Code:    msg,
	}
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
