package pm_autoack

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/k-yomo/pm"
)

// SubscriptionInterceptor automatically ack / nack subscription based on the returned error.
func SubscriptionInterceptor() pm.SubscriptionInterceptor {
	return func(_ *pm.SubscriptionInfo, next pm.MessageHandler) pm.MessageHandler {
		return func(ctx context.Context, m *pubsub.Message) error {
			err := next(ctx, m)
			if err != nil {
				m.Nack()
			} else {
				m.Ack()
			}
			return err
		}
	}
}
