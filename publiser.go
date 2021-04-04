package pm

import (
	"cloud.google.com/go/pubsub"
	"context"
)

// MessagePublisher defines the message publisher invoked by PublishInterceptor to complete the normal
// message publishment.
type MessagePublisher = func(ctx context.Context, topic *pubsub.Topic, m *pubsub.Message) *pubsub.PublishResult
// PublishInterceptor provides a hook to intercept the execution of a publishment.
type PublishInterceptor = func(next MessagePublisher) MessagePublisher


// PublisherConfig represents a configuration of Publisher.
type PublisherConfig struct {
	publishInterceptors []PublishInterceptor
}

// Publisher represents a wrapper of Pub/Sub client focusing on publishment.
type Publisher struct {
	config *PublisherConfig
	*pubsub.Client
}

// NewPublisher initializes new Publisher.
func NewPublisher(pubsubClient *pubsub.Client, opts ...PublisherOption) *Publisher {
	c := PublisherConfig{}
	for _, o := range opts {
		o.apply(&c)
	}
	return &Publisher{
		&c,
		pubsubClient,
	}
}

// Publish publishes Pub/Sub message with applying middlewares
func (p *Publisher) Publish(ctx context.Context, topic *pubsub.Topic, m *pubsub.Message) *pubsub.PublishResult {
	last := publish
	for i := len(p.config.publishInterceptors) - 1; i >= 0; i-- {
		last = p.config.publishInterceptors[i](last)
	}
	return last(ctx, topic, m)
}

func publish(ctx context.Context, topic *pubsub.Topic, m *pubsub.Message) *pubsub.PublishResult {
	return topic.Publish(ctx, m)
}
