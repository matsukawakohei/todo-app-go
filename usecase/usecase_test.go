package usecase

import (
	"context"
	"testing"
)

type requestIDContextKey struct{}

const testRequestID = "test-requiest-id"

func assertRequestID(t *testing.T, ctx context.Context) {
	t.Helper()

	if got, ok := ctx.Value(requestIDContextKey{}).(string); !ok || got != testRequestID {
		t.Errorf("requiest ID want %v; but got %v", testRequestID, got)
	}
}

func makeRequestIDContext() context.Context {
	return context.WithValue(
		context.Background(),
		requestIDContextKey{},
		testRequestID,
	)
}
