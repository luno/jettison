package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/models"
	"github.com/luno/jettison/trace"
)

// UnaryClientInterceptor returns an interceptor that inserts a new hop
// on any jettison error it encounters during unary gRPC calls, and packs any
// jettison metadata in the context into gRPC metadata on the call.
func UnaryClientInterceptor(ctx context.Context, method string,
	req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	err := invoker(outgoingContext(ctx), method, req, reply, cc, opts...)
	return incomingError(err)
}

// StreamClientInterceptor returns an interceptor that inserts a new hop
// on any Jettison error it encounters while creating the client stream, and packs
// any jettison metadata in the context into gRPC metadata on the call. It also
// wraps the returned grpc.ClientStream to support Jettison errors.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	res, err := streamer(outgoingContext(ctx), desc, cc, method, opts...)
	if err != nil {
		return nil, incomingError(err)
	}
	return &clientStream{ClientStream: res}, nil
}

// UnaryServerInterceptor returns an interceptor that unpacks any jettison
// gRPC metadata into the context passed to the server.
func UnaryServerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	interface{}, error,
) {
	a, err := handler(incomingContext(ctx), req)
	return a, outgoingError(err)
}

// StreamServerInterceptor returns an interceptor that unpacks any jettison
// gRPC metadata into the context passed to the server.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream,
	info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	err := handler(srv, &serverStream{ServerStream: ss, ctx: incomingContext(ss.Context())})
	return outgoingError(err)
}

// incomingError converts all non-nil errors into jettison errors.
// If a valid jettison error was sent over the wire a new hop is added
// otherwise the error is wrapped.
func incomingError(err error) error {
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

	je, statusErr := FromStatus(s)
	if errors.Is(statusErr, ErrInvalidError) {
		return errors.Wrap(err, "grpc status error", j.KS("code", s.Code().String()))
	} else if statusErr != nil {
		return errors.Wrap(err, "invalid jettison error", j.KS("err", statusErr.Error()))
	}

	// Push a new hop to the front of the queue.
	h := models.NewHop()
	h.StackTrace = trace.GetStackTraceLegacy(4)
	je.Hops = append([]models.Hop{h}, je.Hops...)
	return errors.Wrap(je, "", errors.WithStackTrace())
}

// outgoingError converts any err into one that will include more details when sent over GRPC
func outgoingError(err error) error {
	if err == nil {
		return nil
	}
	je, ok := err.(*errors.JettisonError)
	if !ok {
		return err
	}
	return gRPCWrap(je)
}

// serverStream is a wrapper of a grpc.ServerStream implementation that decodes
// any jettison options found in the context as jettison options.
type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}

// clientStream is a wrapper of a grpc.ClientStream implementation that
// inserts a new hop on any Jettison error it encounters while streaming.
type clientStream struct {
	grpc.ClientStream
}

func (cs *clientStream) SendMsg(m interface{}) error {
	return incomingError(cs.ClientStream.SendMsg(m))
}

func (cs *clientStream) RecvMsg(m interface{}) error {
	return incomingError(cs.ClientStream.RecvMsg(m))
}
