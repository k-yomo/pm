package pm_recovery

import (
	"cloud.google.com/go/pubsub"
	"context"
	"testing"
)

func TestSubscriptionInterceptor(t *testing.T) {
	t.Parallel()

	next := func(ctx context.Context, m *pubsub.Message) error {
		panic("panic")
	}

	t.Run("recovers with default recovery handler", func(t *testing.T) {
		t.Parallel()
		interceptor := SubscriptionInterceptor()
		_ = interceptor(next)(context.Background(), &pubsub.Message{})
	})

	t.Run("recovers with default recovery handler", func(t *testing.T) {
		t.Parallel()

		var called bool
		opts := []Option{WithRecoveryHandlerContext(func(ctx context.Context, p interface{}) {
			called = true
		})}
		interceptor := SubscriptionInterceptor(opts...)
		_ = interceptor(next)(context.Background(), &pubsub.Message{})
		if !called {
			t.Error("The custom recovery handler is not called")
		}
	})

	t.Run("recovers with debug recovery handler", func(t *testing.T) {
		t.Parallel()

		opts := []Option{WithDebugMode()}
		interceptor := SubscriptionInterceptor(opts...)
		_ = interceptor(next)(context.Background(), &pubsub.Message{})
	})
}
