package pm

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/pkg/errors"
	"log"
)

// MessageHandler defines the message handler invoked by SubscriptionInterceptor to complete the normal
// message handling.
type MessageHandler = func(ctx context.Context, m *pubsub.Message) error

// SubscriptionInterceptor provides a hook to intercept the execution of a message handling.
type SubscriptionInterceptor = func(next MessageHandler) MessageHandler

type subscriberOptions struct {
	subscriptionInterceptors []SubscriptionInterceptor
}

// Subscriber represents a wrapper of Pub/Sub client mainly focusing on pull subscription.
type Subscriber struct {
	opts                 *subscriberOptions
	pubsubClient         *pubsub.Client
	subscriptionHandlers map[string]MessageHandler
	cancel               context.CancelFunc
}

// NewSubscriber initializes new Subscriber.
func NewSubscriber(pubsubClient *pubsub.Client, opt ...SubscriberOption) *Subscriber {
	opts := subscriberOptions{}
	for _, o := range opt {
		o.apply(&opts)
	}
	return &Subscriber{
		opts:                 &opts,
		pubsubClient:         pubsubClient,
		subscriptionHandlers: map[string]MessageHandler{},
	}
}

// HandleSubscriptionFunc registers subscription handler for the given id's subscription.
// If subscription does not exist, it will return error.
func (s *Subscriber) HandleSubscriptionFunc(subscriptionID string, f MessageHandler) error {
	sub := s.pubsubClient.Subscription(subscriptionID)
	ok, err := sub.Exists(context.Background())
	if err != nil {
		return err
	}
	if !ok {
		return errors.Errorf("pubsub subscription '%s' does not exist", subscriptionID)
	}
	s.subscriptionHandlers[subscriptionID] = f

	return nil
}

// Run starts running registered pull subscriptions.
func (s *Subscriber) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	for subscriptionID, f := range s.subscriptionHandlers {
		sub := s.pubsubClient.Subscription(subscriptionID)
		f := f
		go func() {
			last := f
			for i := len(s.opts.subscriptionInterceptors) - 1; i >= 0; i-- {
				last = s.opts.subscriptionInterceptors[i](last)
			}
			err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
				_ = last(ctx, m)
			})
			if err != nil {
				log.Printf("%v\n", err)
			}
		}()
	}
}

// Close closes running subscriptions gracefully.
func (s *Subscriber) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}
