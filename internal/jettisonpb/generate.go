// This file is used to compile the jettisonpb package's proto files.
// Usage: go generate <path to this directory>

//go:generate protoc --go_out=plugins=grpc:. --proto_path=${GOPATH}/src:. ./jettison.proto

package jettisonpb
