package pm

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/pkg/errors"
	"log"
)

type MessageHandler = func(ctx context.Context, m *pubsub.Message) error
type SubscriptionInterceptor = func(next MessageHandler) MessageHandler

type SubscriberConfig struct {
	subscriptionInterceptors []SubscriptionInterceptor
}

type Subscriber struct {
	config               *SubscriberConfig
	pubsubClient         *pubsub.Client
	subscriptionHandlers map[string]MessageHandler
	cancel               context.CancelFunc
}

func NewSubscriber(pubsubClient *pubsub.Client, opts ...SubscriberOption) *Subscriber {
	c := SubscriberConfig{}
	for _, o := range opts {
		o.apply(&c)
	}
	return &Subscriber{
		config:               &c,
		pubsubClient:         pubsubClient,
		subscriptionHandlers: map[string]MessageHandler{},
	}
}

func (p *Subscriber) HandleSubscriptionFunc(subscriptionID string, f MessageHandler) error {
	ok, err := p.pubsubClient.Subscription(subscriptionID).Exists(context.Background())
	if err != nil {
		return err
	}
	if !ok {
		return errors.Errorf("pubsub subscription '%s' does not exist", subscriptionID)
	}
	p.subscriptionHandlers[subscriptionID] = f

	return nil
}

func (p *Subscriber) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	for subscriptionID, f := range p.subscriptionHandlers {
		sub := p.pubsubClient.Subscription(subscriptionID)
		f := f
		go func() {
			last := f
			for i := len(p.config.subscriptionInterceptors) - 1; i >= 0; i-- {
				last = p.config.subscriptionInterceptors[i](last)
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

func (p *Subscriber) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
