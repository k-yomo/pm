package pm

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
)

// Subscriber represents a wrapper of Pub/Sub client mainly focusing on pull subscription.
type Subscriber struct {
	opts                 *subscriberOptions
	pubsubClient         *pubsub.Client
	subscriptionHandlers map[string]*subscriptionHandler
	cancel               context.CancelFunc
}

type subscriptionHandler struct {
	topicID      string
	subscription *pubsub.Subscription
	handleFunc   MessageHandler
}

// MessageHandler defines the message handler invoked by SubscriptionInterceptor to complete the normal
// message handling.
type MessageHandler = func(ctx context.Context, m *pubsub.Message) error

// NewSubscriber initializes new Subscriber.
func NewSubscriber(pubsubClient *pubsub.Client, opt ...SubscriberOption) *Subscriber {
	opts := subscriberOptions{}
	for _, o := range opt {
		o.apply(&opts)
	}
	return &Subscriber{
		opts:                 &opts,
		pubsubClient:         pubsubClient,
		subscriptionHandlers: map[string]*subscriptionHandler{},
	}
}

// HandleSubscriptionFunc registers subscription handler for the given id's subscription.
// If subscription does not exist, it will return error.
func (s *Subscriber) HandleSubscriptionFunc(subscriptionID string, f MessageHandler) error {
	if _, ok := s.subscriptionHandlers[subscriptionID]; ok {
		return fmt.Errorf("handler for subscription '%s' is already registered", subscriptionID)
	}
	sub := s.pubsubClient.Subscription(subscriptionID)
	cfg, err := sub.Config(context.Background())
	if err != nil {
		return err
	}
	s.subscriptionHandlers[subscriptionID] = &subscriptionHandler{
		topicID:      cfg.Topic.ID(),
		subscription: sub,
		handleFunc:   f,
	}

	return nil
}

// HandleSubscriptionFuncMap registers multiple subscription handlers at once.
// This function take map of key[subscription id]: value[corresponding message handler] pairs.
func (s *Subscriber) HandleSubscriptionFuncMap(funcMap map[string]MessageHandler) error {
	eg := errgroup.Group{}
	for subscriptionID, f := range funcMap {
		subscriptionID := subscriptionID
		f := f
		eg.Go(func() error {
			return s.HandleSubscriptionFunc(subscriptionID, f)
		})
	}
	return eg.Wait()
}

// Run starts running registered pull subscriptions.
func (s *Subscriber) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	for subscriptionID, handler := range s.subscriptionHandlers {
		sub := s.pubsubClient.Subscription(subscriptionID)
		h := handler
		subscriptionInfo := SubscriptionInfo{
			TopicID:        h.topicID,
			SubscriptionID: h.subscription.ID(),
		}
		go func() {
			last := h.handleFunc
			for i := len(s.opts.subscriptionInterceptors) - 1; i >= 0; i-- {
				last = s.opts.subscriptionInterceptors[i](&subscriptionInfo, last)
			}
			err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
				_ = last(ctx, m)
			})
			if err != nil {
				log.Printf("%+v\n", err)
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
