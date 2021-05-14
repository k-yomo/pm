package pm_zap

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/k-yomo/pm"
	pm_logging "github.com/k-yomo/pm/middleware/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// SubscriptionInterceptor returns a subscription interceptor that optionally logs the subscription process.
func SubscriptionInterceptor(logger *zap.Logger, opt ...Option) pm.SubscriptionInterceptor {
	opts := &options{
		shouldLog:       pm_logging.DefaultLogDecider,
		messageProducer: DefaultMessageProducer,
		timestampFormat: time.RFC3339,
	}
	for _, o := range opt {
		o.apply(opts)
	}

	return func(info *pm.SubscriptionInfo, next pm.MessageHandler) pm.MessageHandler {
		return func(ctx context.Context, m *pubsub.Message) error {
			startTime := time.Now()
			newCtx := newLoggerForProcess(ctx, logger, info, startTime, opts.timestampFormat)

			err := next(ctxzap.ToContext(newCtx, logger), m)

			if opts.shouldLog(info, err) {
				opts.messageProducer(
					newCtx, fmt.Sprintf("finished processing message '%s'", m.ID),
					err,
					time.Since(startTime),
				)
			}
			return err
		}
	}
}

func newLoggerForProcess(ctx context.Context, logger *zap.Logger, info *pm.SubscriptionInfo, start time.Time, timestampFormat string) context.Context {
	var fields []zapcore.Field
	fields = append(fields, zap.String("pubsub.start_time", start.Format(timestampFormat)))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, zap.String("pubsub.deadline", d.Format(timestampFormat)))
	}
	fields = append(fields, zap.String("pubsub.topic_id", info.TopicID), zap.String("pubsub.subscription_id", info.SubscriptionID))
	return ctxzap.ToContext(ctx, logger.With(fields...))
}
