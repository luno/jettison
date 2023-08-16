package testgrpc

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"github.com/luno/jettison/errors"
	jetgrpc "github.com/luno/jettison/grpc"
	"github.com/luno/jettison/grpc/test/testpb"
)

type Client struct {
	cl   testpb.TestClient
	conn *grpc.ClientConn
}

func NewClient(t *testing.T, addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(jetgrpc.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(jetgrpc.StreamClientInterceptor))
	if err != nil {
		return nil, err
	}

	return &Client{
		cl:   testpb.NewTestClient(conn),
		conn: conn,
	}, nil
}

func (cl *Client) Close() error {
	return cl.conn.Close()
}

func (cl *Client) ClientPB() testpb.TestClient {
	return cl.cl
}

func (cl *Client) ErrorWithCode(ctx context.Context, code string) error {
	_, err := cl.cl.ErrorWithCode(ctx,
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

func (cl *Client) StreamThenError(count int, code string) (int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sc, err := cl.cl.StreamThenError(ctx,
		&testpb.StreamRequest{
			Code:          code,
			ResponseCount: int64(count),
		})
	if err != nil {
		return 0, err
	}

	var empties int
	for i := 0; i < count; i++ {
		_, err := sc.Recv()
		if err != nil {
			return i, errors.Wrap(err, "unexpected error")
		}
		empties++
	}

	// Expect the next call to error with code.
	_, err = sc.Recv()
	if err != nil {
		return empties, err
	}
	return empties + 1, nil
}
