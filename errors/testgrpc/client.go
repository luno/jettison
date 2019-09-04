package testgrpc

import (
	"context"
	"testing"

	"github.com/luno/jettison/errors/testpb"
	"github.com/luno/jettison/interceptors"
	"google.golang.org/grpc"
)

type Client struct {
	cl testpb.TestClient
}

func NewClient(t *testing.T, addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(interceptors.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(interceptors.StreamClientInterceptor))
	if err != nil {
		return nil, err
	}

	return &Client{
		cl: testpb.NewTestClient(conn),
	}, nil
}

func (cl *Client) ErrorWithCode(code string) error {
	_, err := cl.cl.ErrorWithCode(context.Background(),
		&testpb.ErrorWithCodeRequest{
			Code: code,
		})
	return err
}

func (cl *Client) WrapErrorWithCode(code string, wraps int64) error {
	_, err := cl.cl.WrapErrorWithCode(context.Background(),
		&testpb.WrapErrorWithCodeRequest{
			Code:  code,
			Wraps: wraps,
		})
	return err
}
