// Package models contains representations of Jettison objects that are passed
// to loggers.
package models

import (
	"sort"
	"time"
)

type Level string

type Log struct {
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Level     Level     `json:"level"`
	Timestamp time.Time `json:"timestamp"`

	Hops       []Hop      `json:"hops,omitempty"`
	Parameters []KeyValue `json:"parameters,omitempty"`
	ErrorCode  *string    `json:"error_code,omitempty"`
}

// SetKey updates the list of parameters in the log with the given key/value pair.
func (l *Log) SetKey(key, value string) {
	if l == nil {
		return
	}

	l.Parameters = append(l.Parameters, KeyValue{
		Key:   key,
		Value: value,
	})
}

// SetSource updates the source of the log.
func (l *Log) SetSource(src string) {
	if l == nil {
		return
	}

	l.Source = src
}

type Hop struct {
	Binary     string   `json:"binary" protocp:"1"`
	StackTrace []string `json:"stack_trace,omitempty" protocp:"3"`
	Errors     []Error  `json:"errors,omitempty" protocp:"2"`
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

	return *h
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
	if e.Parameters != nil {
		for _, p := range e.Parameters {
			res.Parameters = append(res.Parameters, p)
		}
	}

	return res
}

type KeyValue struct {
	Key   string `json:"key" protocp:"1"`
	Value string `json:"value" protocp:"2"`
}
