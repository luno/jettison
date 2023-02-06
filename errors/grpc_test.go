package errors_test

import (
	"context"
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/errors/testgrpc"
	"github.com/luno/jettison/errors/testpb"
	"github.com/luno/jettison/internal"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/jtest"
)

func TestNewOverGrpc(t *testing.T) {
	l, err := net.Listen("tcp", "")
	require.NoError(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	require.NoError(t, err)
	defer cl.Close()

	errTrue := cl.ErrorWithCode("1")
	require.Error(t, errTrue)

	errFalse := cl.ErrorWithCode("2")
	require.Error(t, errFalse)

	ref := errors.New("reference", j.C("1"))
	assert.True(t, errors.Is(errTrue, ref))
	assert.False(t, errors.Is(errFalse, ref))
}

func TestWrapOverGrpc(t *testing.T) {
	l, err := net.Listen("tcp", "")
	require.NoError(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	require.NoError(t, err)
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
	l, err := net.Listen("tcp", "")
	require.NoError(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	require.NoError(t, err)
	defer cl.Close()

	err = cl.ErrorWithCode("1")
	require.Error(t, err)

	je, ok := err.(*errors.JettisonError)
	require.True(t, ok)
	require.Len(t, je.Hops, 2)

	bb, err := json.MarshalIndent(je.Hops[0].StackTrace, "", "  ")
	require.NoError(t, err)

	expected := `[
  "github.com/luno/jettison/errors/testpb/test.pb.go:220 (*testClient).ErrorWithCode",
  "github.com/luno/jettison/errors/testgrpc/client.go:41",
  "github.com/luno/jettison/errors/grpc_test.go:78 TestClientStacktrace",
  "testing/testing.go:X tRunner",
  "runtime/asm_X.s:X goexit"
]`

	require.Equal(t, expected, string(internal.StripTestStacks(t, bb)))
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
		require.NoError(t, err)
		defer l.Close()

		_, stop := testgrpc.NewServer(t, l)
		defer stop()

		cl, err := testgrpc.NewClient(t, l.Addr().String())
		require.NoError(t, err)
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
	require.NoError(t, err)
	require.NoError(t, l.Close())

	// Nothing is listening
	cl, err := testgrpc.NewClient(t, l.Addr().String())
	require.NoError(t, err)
	defer cl.Close()

	err = cl.ErrorWithCode("")
	require.NotNil(t, err)

	jerr := new(errors.JettisonError)
	require.True(t, errors.As(err, &jerr))

	require.Equal(t, "grpc status error", jerr.Hops[0].Errors[0].Code)
	require.Equal(t, "code", jerr.Hops[0].Errors[0].Parameters[0].Key)
	require.Equal(t, "Unavailable", jerr.Hops[0].Errors[0].Parameters[0].Value)
}

func TestContextCanceled(t *testing.T) {
	l, err := net.Listen("tcp", "")
	require.NoError(t, err)
	defer l.Close()

	_, stop := testgrpc.NewServer(t, l)
	defer stop()

	cl, err := testgrpc.NewClient(t, l.Addr().String())
	require.NoError(t, err)
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
