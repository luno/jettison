package test_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/go-stack/stack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/grpc/test/testgrpc"
	"github.com/luno/jettison/grpc/test/testpb"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/jtest"
	"github.com/luno/jettison/log"
	"github.com/luno/jettison/trace"
)

func TestNewOverGrpc(t *testing.T) {
	l, err := net.Listen("tcp", "")
	jtest.RequireNil(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	jtest.RequireNil(t, err)
	defer cl.Close()

	ctx := context.Background()
	errTrue := cl.ErrorWithCode(ctx, "1")
	require.Error(t, errTrue)

	errFalse := cl.ErrorWithCode(ctx, "2")
	require.Error(t, errFalse)

	ref := errors.New("reference", j.C("1"))
	assert.True(t, errors.Is(errTrue, ref))
	assert.False(t, errors.Is(errFalse, ref))
}

func TestWrapOverGrpc(t *testing.T) {
	l, err := net.Listen("tcp", "")
	jtest.RequireNil(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	jtest.RequireNil(t, err)
	defer cl.Close()

	errTrue := cl.WrapErrorWithCode("1", 10)
	require.Error(t, errTrue)

	errFalse := cl.WrapErrorWithCode("2", 10)
	require.Error(t, errFalse)

	ref := errors.New("reference", j.C("1"))
	assert.True(t, errors.Is(errTrue, ref))
	assert.False(t, errors.Is(errFalse, ref))
}

func TestClientStacktrace(t *testing.T) {
	errors.SetTraceConfigTesting(t, trace.StackConfig{
		TrimRuntime: true,
		Format: func(call stack.Call) string {
			return fmt.Sprintf("%s %n", call, call)
		},
	})
	l, err := net.Listen("tcp", "")
	jtest.RequireNil(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	jtest.RequireNil(t, err)
	defer cl.Close()

	err = cl.ErrorWithCode(context.Background(), "1")
	require.Error(t, err)
	je := err.(*errors.JettisonError)

	exp := []string{
		"interceptors.go incomingError",
		"interceptors.go UnaryClientInterceptor",
		"call.go (*ClientConn).Invoke",
		"test.pb.go (*testClient).ErrorWithCode",
		"client.go (*Client).ErrorWithCode",
		"grpc_test.go TestClientStacktrace",
	}
	assert.Equal(t, exp, je.StackTrace)
}

func TestStreamThenError(t *testing.T) {
	tests := []struct {
		Name  string
		Count int
	}{
		{
			Name:  "zero",
			Count: 0,
		}, {
			Name:  "ten",
			Count: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, makeTestStreamWithError(test.Name, test.Count))
	}
}

func makeTestStreamWithError(name string, count int) func(t *testing.T) {
	return func(t *testing.T) {
		l, err := net.Listen("tcp", "")
		jtest.RequireNil(t, err)
		defer l.Close()

		_, stop := testgrpc.NewServer(t, l)
		defer stop()

		cl, err := testgrpc.NewClient(t, l.Addr().String())
		jtest.RequireNil(t, err)
		defer cl.Close()

		c, err := cl.StreamThenError(count, name)
		require.Equal(t, c, count, "unexpected ", "count mismatch, error: %v", err)

		ref := errors.New("reference", j.C(name))
		assert.True(t, errors.Is(err, ref))
	}
}

func TestWrappingGrpcError(t *testing.T) {
	// Get an open port
	l, err := net.Listen("tcp", "")
	jtest.RequireNil(t, err)
	jtest.RequireNil(t, l.Close())

	// Nothing is listening
	cl, err := testgrpc.NewClient(t, l.Addr().String())
	jtest.RequireNil(t, err)
	defer cl.Close()

	ctx := context.Background()
	err = cl.ErrorWithCode(ctx, "")
	require.NotNil(t, err)

	assert.Equal(t, "grpc status error", err.Error())
	assert.Equal(t, []string{"grpc status error"}, errors.GetCodes(err))

	kvs := errors.GetKeyValues(err)
	assert.Equal(t, "Unavailable", kvs["code"])
}

func TestDeadlineExceededClient(t *testing.T) {
	l, err := net.Listen("tcp", "")
	jtest.RequireNil(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	jtest.RequireNil(t, err)
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	<-ctx.Done()

	err = cl.ErrorWithCode(ctx, "")
	require.NotNil(t, err)

	jtest.Assert(t, context.DeadlineExceeded, err)
	assert.Equal(t, "grpc status error", err.Error())
	assert.Equal(t, []string{"grpc status error"}, errors.GetCodes(err))

	kvs := errors.GetKeyValues(err)
	assert.Equal(t, "DeadlineExceeded", kvs["code"])
}

func TestContextCanceled(t *testing.T) {
	l, err := net.Listen("tcp", "")
	jtest.RequireNil(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	jtest.RequireNil(t, err)
	defer cl.Close()

	ctx, cancel := context.WithCancel(context.Background())

	sc, err := cl.ClientPB().StreamThenError(ctx, &testpb.StreamRequest{ResponseCount: 100000})
	jtest.RequireNil(t, err)

	_, err = sc.Recv()
	jtest.RequireNil(t, err)

	cancel()

	for {
		//
		_, err := sc.Recv()
		if err == nil {
			// Gobble buffered responses
			continue
		}
		jtest.Require(t, context.Canceled, err)
		break
	}
}

func TestContextKeys(t *testing.T) {
	l, err := net.Listen("tcp", "")
	jtest.RequireNil(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	jtest.RequireNil(t, err)
	defer cl.Close()

	ctx := log.ContextWith(context.Background(), j.KV("09%-_MANYproblems", "hello"))
	err = cl.ErrorWithCode(ctx, "CODE1234")

	codes := errors.GetCodes(err)
	assert.Equal(t, []string{"CODE1234"}, codes)
}
