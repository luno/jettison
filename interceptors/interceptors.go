package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/models"
)

// UnaryClientInterceptor returns an interceptor that inserts a new hop
// on any jettison error it encounters during unary gRPC calls, and packs any
// jettison metadata in the context into gRPC metadata on the call.
func UnaryClientInterceptor(ctx context.Context, method string,
	req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	err := invoker(withMetadata(ctx), method, req, reply, cc, opts...)
	return intercept(err)
}

// StreamClientInterceptor returns an interceptor that inserts a new hop
// on any Jettison error it encounters while creating the client stream, and packs
// any jettison metadata in the context into gRPC metadata on the call. It also
// wraps the returned grpc.ClientStream to support Jettison errors.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	res, err := streamer(withMetadata(ctx), desc, cc, method, opts...)
	if err != nil {
		return nil, intercept(err)
	}

	return &clientStream{ClientStream: res}, nil
}

// UnaryServerInterceptor returns an interceptor that unpacks any jettison
// gRPC metadata into the context passed to the server.
func UnaryServerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	interface{}, error,
) {
	return handler(fromMetadata(ctx), req)
}

// StreamServerInterceptor returns an interceptor that unpacks any jettison
// gRPC metadata into the context passed to the server.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
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

// intercept converts all non-nil errors into jettison errors. If a valid jettison error was sent over the wire
// a new hop is added otherwise the error is wrapped.
func intercept(err error) error {
	if err == nil {
		return nil
	}

	s, ok := status.FromError(err)
	if !ok {
		return errors.Wrap(err, "non-grpc error")
	}

	if s.Code() == codes.Canceled {
		return errors.Wrap(context.Canceled, "grpc error")
	} else if s.Code() == codes.DeadlineExceeded {
		return errors.Wrap(context.DeadlineExceeded, "grpc error")
	}

	je, statusErr := errors.FromStatus(s)
	if errors.Is(statusErr, errors.ErrInvalidError) {
		return errors.Wrap(err, "grpc status error", j.KS("code", s.Code().String()))
	} else if statusErr != nil {
		return errors.Wrap(err, "invalid jettison error", j.KS("err", statusErr.Error()))
	}

	// Push a new hop to the front of the queue.
	h := internal.NewHop()
	h.StackTrace = internal.GetStackTrace(4)
	je.Hops = append([]models.Hop{h}, je.Hops...)
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

// clientStream is a wrapper of a grpc.ClientStream implementation that
// inserts a new hop on any Jettison error it encounters while streaming.
type clientStream struct {
	grpc.ClientStream
}

func (cs *clientStream) SendMsg(m interface{}) error {
	return intercept(cs.ClientStream.SendMsg(m))
}

func (cs *clientStream) RecvMsg(m interface{}) error {
	return intercept(cs.ClientStream.RecvMsg(m))
}
