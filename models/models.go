// Package models contains representations of Jettison objects that are passed
// to loggers.
package models

import (
	"sort"
)

// Metadata is the extra info available at each level of the error tree
type Metadata struct {
	// Trace is info on the source of the error
	Trace Hop `json:"trace" protocp:"1"`
	// Code is an identifier for the type of error
	Code string `json:"code" protocp:"2"`
	// KV is a list of extra info in the error
	KV []KeyValue `json:"kv" protocp:"3"`
}

func (m *Metadata) IsZero() bool {
	return m.Trace.IsZero() && m.Code == "" && len(m.KV) == 0
}

func (m *Metadata) SetKey(key, value string) {
	if m == nil {
		return
	}
	m.KV = append(m.KV, KeyValue{
		Key:   key,
		Value: value,
	})
}

// SetSource updates the source of the most recently added error in the hop.
func (m *Metadata) SetSource(string) {}

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

type KeyValue struct {
	Key   string `json:"key" protocp:"1"`
	Value string `json:"value" protocp:"2"`
}
