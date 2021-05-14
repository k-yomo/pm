package pm_zap

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/k-yomo/pm/middleware/logging"
	"go.uber.org/zap"
	"time"
)

type options struct {
	shouldLog       pm_logging.LogDecider
	messageProducer MessageProducer
	timestampFormat string
}

// MessageProducer produces a user defined log message
type MessageProducer func(ctx context.Context, msg string, err error, duration time.Duration)

// DefaultMessageProducer writes the default message
func DefaultMessageProducer(ctx context.Context, msg string, err error, duration time.Duration) {
	logger := ctxzap.Extract(ctx)
	durationField := zap.Float32("pubsub.time_ms", durationToMilliseconds(duration))
	if err != nil {
		logger.Error(msg, zap.Error(err), durationField)
	} else {
		logger.Info(msg, durationField)
	}
}

// Option is a option to change configuration.
type Option interface {
	apply(*options)
}

type OptionFunc struct {
	f func(*options)
}

func (s *OptionFunc) apply(so *options) {
	s.f(so)
}

func newOptionFunc(f func(*options)) *OptionFunc {
	return &OptionFunc{
		f: f,
	}
}

// WithLogDecider customizes the function for deciding if the pm interceptor should log.
func WithLogDecider(f pm_logging.LogDecider) Option {
	return newOptionFunc(func(o *options) {
		o.shouldLog = f
	})
}

// WithMessageProducer customizes the function for logging.
func WithMessageProducer(f MessageProducer) Option {
	return newOptionFunc(func(o *options) {
		o.messageProducer = f
	})
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}
