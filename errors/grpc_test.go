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

	_ = testgrpc.NewServer(t, l)
	cl, err := testgrpc.NewClient(t, l.Addr().String())
	require.NoError(t, err)

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

	_ = testgrpc.NewServer(t, l)
	cl, err := testgrpc.NewClient(t, l.Addr().String())
	require.NoError(t, err)

	errTrue := cl.WrapErrorWithCode("1", 10)
	require.Error(t, errTrue)

	errFalse := cl.WrapErrorWithCode("2", 10)
	require.Error(t, errFalse)

	ref := errors.New("reference", j.C("1"))
	assert.True(t, errors.Is(errTrue, ref))
	assert.False(t, errors.Is(errFalse, ref))
}
