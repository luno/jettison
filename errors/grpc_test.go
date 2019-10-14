package errors_test

import (
	"net"
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/errors/testgrpc"
	"github.com/luno/jettison/j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
