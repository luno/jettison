package internal

//go:generate protocp --pkg=github.com/luno/jettison/models --proto_pkg=github.com/luno/jettison/internal/jettisonpb --output=models.protocp.go --register=src/bitx/backends/protocp/base_converters.json Hop Error KeyValue
