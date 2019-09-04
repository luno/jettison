package testgrpc

import (
	"context"
	"net"
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/errors/testpb"
	"github.com/luno/jettison/interceptors"
	"github.com/luno/jettison/j"
	"google.golang.org/grpc"
)

type Server struct{}

func NewServer(t *testing.T, l net.Listener) *Server {
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.UnaryServerInterceptor),
		grpc.StreamInterceptor(interceptors.StreamServerInterceptor))

	srv := new(Server)
	testpb.RegisterTestServer(grpcSrv, srv)

	go func() {
		err := grpcSrv.Serve(l)
		if err != nil {
			panic(err)
		}
	}()

	return srv
}

func (srv *Server) ErrorWithCode(ctx context.Context,
	req *testpb.ErrorWithCodeRequest) (*testpb.Empty, error) {

	return nil, errors.New("error with code", j.C(req.Code))
}

func (srv *Server) WrapErrorWithCode(ctx context.Context,
	req *testpb.WrapErrorWithCodeRequest) (*testpb.Empty, error) {

	err := errors.New("wrap error with code", j.C(req.Code))
	for i := int64(0); i < req.Wraps; i++ {
		err = errors.Wrap(err, "wrap")
	}

	return nil, err
}
