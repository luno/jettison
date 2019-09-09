package interceptors

import (
	"context"
	"log"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryClientInterceptor returns an interceptor that inserts a new hop
// on any jettison error it encounters during unary gRPC calls, and packs any
// jettison metadata in the context into gRPC metadata on the call.
func UnaryClientInterceptor(ctx context.Context, method string,
	req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	err := invoker(withMetadata(ctx), method, req, reply, cc, opts...)
	return withNewHop(err)
}

// StreamClientInterceptor returns an interceptor that inserts a new hop
// on any Jettison error it encounters while creating gRPC streams, and packs
// any jettison metadata in the context into gRPC metadata on the call.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {

	res, err := streamer(withMetadata(ctx), desc, cc, method, opts...)
	return res, withNewHop(err)
}

// UnaryServerInterceptor returns an interceptor that unpacks any jettison
// gRPC metadata into the context passed to the server.
func UnaryServerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	interface{}, error) {

	return handler(fromMetadata(ctx), req)
}

// StreamServerInterceptor returns an interceptor that unpacks any jettison
// gRPC metadata into the context passed to the server.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	return handler(srv, &serverStream{ServerStream: ss})
}

// withMetadata returns a new context with any jettison options found in the
// original context encoded as gRPC metadata.
func withMetadata(ctx context.Context) context.Context {
	opts := internal.ContextOptions(ctx)
	if len(opts) == 0 {
		return ctx // nothing to encode as gRPC metadata
	}

	var cd internal.ContextDetails
	if ctxMd, ok := metadata.FromOutgoingContext(ctx); ok {
		cd = internal.ContextDetails(ctxMd.Copy())
	} else {
		cd = make(internal.ContextDetails)
	}
	for _, o := range opts {
		o.Apply(&cd)
	}

	return metadata.NewOutgoingContext(ctx, cd.ToGrpcMetadata())
}

// fromMetadata returns a new context with any jettison gRPC metadata found
// in the original context packed as jettison options.
func fromMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx // no metadata to decode
	}

	opts := internal.FromGrpcMetadata(md)
	for _, o := range opts {
		ctx = internal.ContextWith(ctx, o)
	}

	return ctx
}

func withNewHop(err error) error {
	if err == nil {
		return nil
	}

	status, ok := status.FromError(err)
	if !ok {
		return err
	}

	je, statusErr := errors.FromStatus(status)
	if statusErr != nil {
		log.Printf("jettison/interceptors: Error converting grpc status: %v", err)
		return err
	}
	if len(je.Hops) == 0 {
		return err
	}

	// Push a new hop to the front of the queue.
	je.Hops = append([]models.Hop{internal.NewHop()}, je.Hops...)
	return je
}

// serverStream is a wrapper of a grpc.ServerStream implementation that decodes
// any jettison options found in the context as jettison options.
type serverStream struct {
	grpc.ServerStream
}

func (ss *serverStream) Context() context.Context {
	return fromMetadata(ss.ServerStream.Context())
}
