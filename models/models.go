// Package models contains representations of Jettison objects that are passed
// to loggers.
package models

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
