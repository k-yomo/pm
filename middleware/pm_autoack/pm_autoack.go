package pm_autoack

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/k-yomo/pm"
)

func SubscriptionInterceptor(next pm.MessageHandler) pm.MessageHandler {
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
