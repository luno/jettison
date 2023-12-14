package log

import (
	"time"

	"github.com/luno/jettison/models"
)

type Level string

type ErrorObject struct {
	Code       string             `json:"code"`
	Source     string             `json:"source"`
	Message    string             `json:"message"`
	Stack      []string           `json:"stack,omitempty"`
	StackTrace ElasticStringArray `json:"stacktrace,omitempty"`
	Parameters []models.KeyValue  `json:"parameters,omitempty"`
}

type Entry struct {
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Level     Level     `json:"level"`
	Timestamp time.Time `json:"timestamp"`

	Parameters []models.KeyValue `json:"parameters,omitempty"`
	ErrorCode  *string           `json:"error_code,omitempty"`

	ErrorObject  *ErrorObject  `json:"error_object,omitempty"`
	ErrorObjects []ErrorObject `json:"error_objects,omitempty"`
}

// SetKey updates the list of parameters in the log with the given key/value pair.
func (l *Entry) SetKey(key, value string) {
	if l == nil {
		return
	}

	l.Parameters = append(l.Parameters, models.KeyValue{
		Key:   key,
		Value: value,
	})
}

// SetSource updates the source of the log.
func (l *Entry) SetSource(src string) {
	if l == nil {
		return
	}

	l.Source = src
}
