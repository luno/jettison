package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/errors"
)

// UnaryClientInterceptor intercepts errors, de-serialising any
// // WrappedErrors we find and unpacking any context jettison key-values.
func UnaryClientInterceptor(ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	err := invoker(outgoingContext(ctx), method, req, reply, cc, opts...)
	return incomingError(err)
}

// StreamClientInterceptor intercepts errors, de-serialising any
// WrappedErrors we find and unpacking any context jettison key-values.
func StreamClientInterceptor(ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	res, err := streamer(outgoingContext(ctx), desc, cc, method, opts...)
	if err != nil {
		return nil, incomingError(err)
	}
	return &clientStream{ClientStream: res}, nil
}

// UnaryServerInterceptor intercepts errors, de-serialising any
// WrappedErrors we find and unpacking any context jettison key-values.
func UnaryServerInterceptor(ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	a, err := handler(incomingContext(ctx), req)
	return a, outgoingError(err)
}

// StreamServerInterceptor intercepts errors, de-serialising any
// WrappedErrors we find and unpacking any context jettison key-values.
func StreamServerInterceptor(
	srv any,
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	err := handler(srv, &serverStream{ServerStream: ss, ctx: incomingContext(ss.Context())})
	return outgoingError(err)
}

// incomingError converts all non-nil errors into jettison errors.
// a new stack trace is added representing the stack in this new binary.
func incomingError(err error) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if !ok {
		// Another interceptor may have already converted this error
		return errors.Wrap(err, "", errors.WithStackTrace())
	}
	return errors.Wrap(FromStatus(s), "", errors.WithStackTrace())
}

// outgoingError converts any err into one that will include more details when sent over GRPC
func outgoingError(err error) error {
	if err == nil {
		return nil
	}
	return Wrap(err)
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}

type clientStream struct {
	grpc.ClientStream
}

func (cs *clientStream) SendMsg(m interface{}) error {
	return incomingError(cs.ClientStream.SendMsg(m))
}

func (cs *clientStream) RecvMsg(m interface{}) error {
	return incomingError(cs.ClientStream.RecvMsg(m))
}
