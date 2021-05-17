package pm_logrus

import (
	"context"
	"github.com/k-yomo/pm"
	"testing"
	"time"
)

func TestWithLogDecider(t *testing.T) {
	t.Parallel()

	opts := options{}
	isDecided := false
	customDecider := func(info *pm.SubscriptionInfo, err error) bool {
		isDecided = true
		return true
	}
	WithLogDecider(customDecider).apply(&opts)
	opts.shouldLog(nil, nil)
	if !isDecided {
		t.Errorf("WithLogDecider() is expected to set custom diceider, but it was not set")
	}
}

func TestWithMessageProducer(t *testing.T) {
	t.Parallel()

	opts := options{}

	isProduced := false
	customMessageProducer := func(ctx context.Context, msg string, err error, duration time.Duration) {
		isProduced = true
	}
	WithMessageProducer(customMessageProducer).apply(&opts)
	opts.messageProducer(nil, "", nil, 0)
	if !isProduced {
		t.Errorf("WithMessageProducer() is expected to set custom message producer, but it was not set")
	}
}
