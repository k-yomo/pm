package pm_logrus

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"github.com/k-yomo/pm"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
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

			logger, hook := test.NewNullLogger()
			intercepter := SubscriptionInterceptor(logger)
			callHandler(intercepter(testSubInfo, successMessageHandler))

			if got := len(hook.Entries); got != 1 {
				t.Fatalf("Only 1 log is expected to be emitted, got: %v, want: %v", got, 1)
			}
			entry := hook.Entries[0]
			if entry.Level != logrus.InfoLevel {
				t.Errorf("INFO log is expected to be emitted, got: %v, want: %v", entry.Level, logrus.InfoLevel)
			}
			wantMessage := "finished processing message 'message-id'"
			if entry.Message != wantMessage {
				t.Errorf("INFO log is expected to be emitted, got: %v, want: %v", entry.Message, wantMessage)
			}
		})

		t.Run("Emit error log when processing is successful", func(t *testing.T) {
			t.Parallel()

			logger, hook := test.NewNullLogger()
			intercepter := SubscriptionInterceptor(logger)
			callHandler(intercepter(testSubInfo, failureMessageHandler))

			if got := len(hook.Entries); got != 1 {
				t.Fatalf("Only 1 log is expected to be emitted, got: %v, want: %v", got, 1)
			}
			entry := hook.Entries[0]
			if entry.Level != logrus.ErrorLevel {
				t.Errorf("INFO log is expected to be emitted, got: %v, want: %v", entry.Level, logrus.ErrorLevel)
			}
			wantMessage := "finished processing message 'message-id'"
			if entry.Message != wantMessage {
				t.Errorf("INFO log is expected to be emitted, got: %v, want: %v", entry.Message, wantMessage)
			}
		})
	})

	t.Run("with custom options", func(t *testing.T) {
		t.Run("custom options are applied", func(t *testing.T) {
			t.Parallel()

			logger, hook := test.NewNullLogger()
			intercepter := SubscriptionInterceptor(logger, WithLogDecider(func(info *pm.SubscriptionInfo, err error) bool {
				return false
			}))
			callHandler(intercepter(testSubInfo, successMessageHandler))

			if len(hook.Entries) != 0 {
				t.Errorf("log is not expected to be emitted")
			}
		})
	})
}
