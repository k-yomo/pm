package pm

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/pkg/errors"
	"log"
)

type MessageHandler = func(ctx context.Context, message *pubsub.Message)
type SubscriptionInterceptor = func(next MessageHandler) MessageHandler

type Config struct {
	subscriptionInterceptors []SubscriptionInterceptor
}

type PubSubManager struct {
	config               *Config
	pubsubClient         *pubsub.Client
	subscriptionHandlers map[string]MessageHandler
	cancel               context.CancelFunc
}

func NewPubSubManager(pubsubClient *pubsub.Client, opts ...Option) *PubSubManager {
	c := Config{}
	for _, o := range opts {
		o.apply(&c)
	}
	return &PubSubManager{
		config:       &c,
		pubsubClient: pubsubClient,
		subscriptionHandlers: map[string]MessageHandler{},
	}
}

func (p *PubSubManager) HandleSubscriptionFunc(subscriptionID string, f MessageHandler) error {
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

func (p *PubSubManager) Run() {
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
			if err := sub.Receive(ctx, last); err != nil {
				log.Printf("%v\n", err)
			}
		}()
	}
}

func (p *PubSubManager) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
