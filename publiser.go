package pm

import (
	"cloud.google.com/go/pubsub"
	"context"
)

type MessagePublisher = func(ctx context.Context, topic *pubsub.Topic, m *pubsub.Message) *pubsub.PublishResult
type PublishInterceptor = func(next MessagePublisher) MessagePublisher

type PublisherConfig struct {
	publishInterceptors []PublishInterceptor
}

type Publisher struct {
	config *PublisherConfig
	*pubsub.Client
}

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
