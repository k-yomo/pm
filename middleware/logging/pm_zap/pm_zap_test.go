package pm_zap

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"github.com/k-yomo/pm"
	"go.uber.org/zap"
	zapobserver "go.uber.org/zap/zaptest/observer"
	"testing"
)

func TestSubscriptionInterceptor(t *testing.T) {
	t.Parallel()

	testSubInfo := &pm.SubscriptionInfo{
		TopicID:        "test-topic",
		SubscriptionID: "test-sub",
	}

	successMessageHandler := func(ctx context.Context, m *pubsub.Message) error {
		return nil
	}
	failureMessageHandler := func(ctx context.Context, m *pubsub.Message) error {
		return errors.New("error")
	}

	callHandler := func(f pm.MessageHandler) {
		_ = f(context.Background(), &pubsub.Message{ID: "message-id"})
	}

	t.Run("with default options", func(t *testing.T) {
		t.Run("emit info log when processing is successful", func(t *testing.T) {
			t.Parallel()

			core, obs := zapobserver.New(zap.InfoLevel)
			logger := zap.New(core)

			intercepter := SubscriptionInterceptor(logger)
			callHandler(intercepter(testSubInfo, successMessageHandler))

			if obs.Len() != 1 {
				t.Errorf("INFO log is expected to be emitted")
			}
		})

		t.Run("Emit error log when processing is successful", func(t *testing.T) {
			t.Parallel()

			core, obs := zapobserver.New(zap.ErrorLevel)
			logger := zap.New(core)

			intercepter := SubscriptionInterceptor(logger)
			callHandler(intercepter(testSubInfo, failureMessageHandler))

			if obs.Len() != 1 {
				t.Errorf("ERROR log is expected to be emitted")
			}
		})
	})

	t.Run("with custom options", func(t *testing.T) {
		t.Run("custom options are applied", func(t *testing.T) {
			t.Parallel()

			core, obs := zapobserver.New(zap.DebugLevel)
			logger := zap.New(core)

			intercepter := SubscriptionInterceptor(logger, WithLogDecider(func(info *pm.SubscriptionInfo, err error) bool {
				return false
			}))
			callHandler(intercepter(testSubInfo, successMessageHandler))

			if obs.Len() != 0 {
				t.Errorf("log is not expected to be emitted")
			}
		})
	})
}
