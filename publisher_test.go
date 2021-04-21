package pm

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type publisherOptionForTest struct {
}

func (s publisherOptionForTest) apply(so *publisherOptions) {
	so.publishInterceptors = []PublishInterceptor{nil}
}

func TestNewPublisher(t *testing.T) {
	t.Parallel()

	type args struct {
		pubsubClient *pubsub.Client
		opt          []PublisherOption
	}
	tests := []struct {
		name string
		args args
		want *Publisher
	}{
		{
			name: "Initialize new subscriber",
			args: args{
				pubsubClient: &pubsub.Client{},
				opt:          []PublisherOption{publisherOptionForTest{}},
			},
			want: &Publisher{
				opts:   &publisherOptions{publishInterceptors: []PublishInterceptor{nil}},
				Client: &pubsub.Client{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPublisher(tt.args.pubsubClient, tt.args.opt...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPublisher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublisher_Publish(t *testing.T) {
	t.Parallel()

	pubsubClient, err := pubsub.NewClient(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}

	topic, err := pubsubClient.CreateTopic(context.Background(), fmt.Sprintf("TestPublisher_Publish_%d", time.Now().Unix()))
	if err != nil {
		t.Fatal(err)
	}

	sub, err := pubsubClient.CreateSubscription(
		context.Background(),
		fmt.Sprintf("TestPublisher_Publish_%d", time.Now().Unix()),
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx   context.Context
		topic *pubsub.Topic
		m     *pubsub.Message
	}
	tests := []struct {
		name      string
		publisher *Publisher
		args      args
		wantData  string
		wantErr   bool
	}{
		{
			name:      "Publish message",
			publisher: NewPublisher(pubsubClient),
			args:      args{ctx: context.Background(), topic: topic, m: &pubsub.Message{Data: []byte("test")}},
			wantData:  "test",
		},
		{
			name: "Publish message with interceptors",
			publisher: NewPublisher(pubsubClient, WithPublishInterceptor(func(next MessagePublisher) MessagePublisher {
				return func(ctx context.Context, topic *pubsub.Topic, m *pubsub.Message) *pubsub.PublishResult {
					m.Data = []byte("overwritten by first interceptor")
					return next(ctx, topic, m)
				}
			}, func(next MessagePublisher) MessagePublisher {
				return func(ctx context.Context, topic *pubsub.Topic, m *pubsub.Message) *pubsub.PublishResult {
					m.Data = []byte("overwritten by last interceptor")
					return next(ctx, topic, m)
				}
			})),
			args:     args{ctx: context.Background(), topic: topic, m: &pubsub.Message{Data: []byte("test")}},
			wantData: "overwritten by last interceptor",
		},
		{
			name:      "Publish empty message will result in error",
			publisher: NewPublisher(pubsubClient),
			args:      args{ctx: context.Background(), topic: topic, m: &pubsub.Message{Data: nil}},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.publisher.Publish(tt.args.ctx, tt.args.topic, tt.args.m).Get(tt.args.ctx); !reflect.DeepEqual(err != nil, tt.wantErr) {
				t.Errorf("Publish() = %v, want %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				wg := sync.WaitGroup{}
				wg.Add(1)
				ctx, cancel := context.WithTimeout(tt.args.ctx, 3*time.Second)
				go func() {
					defer wg.Done()
					err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
						m.Ack()
						if string(m.Data) != tt.wantData {
							t.Errorf("publish() gotData = %v, want %v", string(m.Data), tt.wantData)
						}
					})
					if err != nil {
						t.Fatal(err)
					}
				}()
				wg.Wait()
				cancel()
			}
		})
	}

	NewSubscriber(pubsubClient)
}
