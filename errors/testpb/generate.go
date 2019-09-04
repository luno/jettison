//go:generate protoc --go_out=plugins=grpc:. --proto_path=${GOPATH}/src:. ./test.proto
package testpb
