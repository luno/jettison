package interceptors

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luno/jettison/jtest"
)

func TestErrIntercept(t *testing.T) {
	testCases := []struct {
		name    string
		testErr error
		expErr  error
	}{
		{name: "nil is nil"},
		{
			name:    "grpc status canceled gets context error",
			testErr: status.Error(codes.Canceled, ""),
			expErr:  context.Canceled,
		},
		{
			name:    "grpc status deadline exceeded gets context error",
			testErr: status.Error(codes.DeadlineExceeded, ""),
			expErr:  context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := incomingError(tc.testErr)
			jtest.Require(t, tc.expErr, err)
		})
	}
}
