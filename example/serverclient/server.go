package serverclient

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/example/examplepb"
	"github.com/luno/jettison/interceptors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

var _ examplepb.HopperServer = (*Server)(nil)

type Server struct {
	url string
	srv *grpc.Server

	client *Client
}

// Hop bounces a request between this server and it's client for a number of
// hops, and then errors. This illustrates how errors are segmented by gRPC
// call in Jettison.
func (srv *Server) Hop(ctx context.Context, req *examplepb.HopRequest) (
	*examplepb.Empty, error,
) {
	log.Info(ctx, "serverclient: Received a Hop request from a client")

	if req.Hops <= 0 {
		return nil, errors.New("serverclient: run out of hops",
			j.KV("hops", req.Hops))
	}

	if err := srv.client.Hop(ctx, req.Hops); err != nil {
		return nil, errors.Wrap(err, "serverclient: error hopping",
			j.KV("hops", req.Hops))
	}

	return &examplepb.Empty{}, nil
}

func (srv *Server) SetClient(client *Client) {
	srv.client = client
}

func (srv *Server) GetURL() string {
	return srv.url
}

func (srv *Server) Stop() {
	srv.srv.Stop()
}

func NewServer() *Server {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(errors.Wrap(err, "serverclient: net.Listen error"))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.UnaryServerInterceptor),
		grpc.StreamInterceptor(interceptors.StreamServerInterceptor))
	srv := &Server{
		url: l.Addr().String(),
		srv: grpcServer,
	}
	examplepb.RegisterHopperServer(grpcServer, srv)

	go func() {
		err := grpcServer.Serve(l)
		if err != nil {
			panic(errors.Wrap(err, "serverclient: grpcServer.Server error"))
		}
	}()

	return srv
}
