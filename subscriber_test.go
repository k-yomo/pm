package pm

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type subscriberOptionForTest struct {
}

func (s subscriberOptionForTest) apply(so *subscriberOptions) {
	so.subscriptionInterceptors = []SubscriptionInterceptor{nil}
}

func TestNewSubscriber(t *testing.T) {
	type args struct {
		pubsubClient *pubsub.Client
		opt          []SubscriberOption
	}
	tests := []struct {
		name string
		args args
		want *Subscriber
	}{
		{
			name: "Initialize new subscriber",
			args: args{
				pubsubClient: &pubsub.Client{},
				opt:          []SubscriberOption{subscriberOptionForTest{}},
			},
			want: &Subscriber{
				opts:                 &subscriberOptions{subscriptionInterceptors: []SubscriptionInterceptor{nil}},
				pubsubClient:         &pubsub.Client{},
				subscriptionHandlers: map[string]MessageHandler{},
				cancel:               nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSubscriber(tt.args.pubsubClient, tt.args.opt...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSubscriber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscriber_HandleSubscriptionFunc(t *testing.T) {
	t.Parallel()

	pubsubClient, err := pubsub.NewClient(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}

	topic, err := pubsubClient.CreateTopic(context.Background(), fmt.Sprintf("TestNewSubscriber_%d", time.Now().Unix()))
	if err != nil {
		t.Fatal(err)
	}

	sub, err := pubsubClient.CreateSubscription(
		context.Background(),
		fmt.Sprintf("TestNewSubscriber_%d", time.Now().Unix()),
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		subscriptionID string
		f              MessageHandler
	}
	tests := []struct {
		name       string
		subscriber *Subscriber
		args       args
		wantErr    bool
	}{
		{
			name:       "sets subscription handler",
			subscriber: NewSubscriber(pubsubClient),
			args: args{
				subscriptionID: sub.ID(),
				f: func(ctx context.Context, m *pubsub.Message) error {
					return nil
				},
			},
		},
		{
			name:       "if the subscription does not exist, it returns error",
			subscriber: NewSubscriber(pubsubClient),
			args: args{
				subscriptionID: "invalid",
				f:              nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.subscriber.HandleSubscriptionFunc(tt.args.subscriptionID, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("HandleSubscriptionFunc() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, ok := tt.subscriber.subscriptionHandlers[tt.args.subscriptionID]; ok == tt.wantErr {
				t.Errorf("HandleSubscriptionFunc() set: %v, want set: %v", ok, !tt.wantErr)
			}
		})
	}
}

func TestSubscriber_Run(t *testing.T) {
	t.Parallel()

	pubsubClient, err := pubsub.NewClient(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}

	topic, err := pubsubClient.CreateTopic(context.Background(), fmt.Sprintf("TestSubscriber_Run_%d", time.Now().Unix()))
	if err != nil {
		t.Fatal(err)
	}

	sub, err := pubsubClient.CreateSubscription(
		context.Background(),
		fmt.Sprintf("TestSubscriber_Run_%d", time.Now().Unix()),
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err != nil {
		t.Fatal(err)
	}

	subscriber := NewSubscriber(
		pubsubClient,
		WithSubscriptionInterceptor(func(next MessageHandler) MessageHandler {
			return func(ctx context.Context, m *pubsub.Message) error {
				m.Attributes = map[string]string{"intercepted": "abc"}
				return next(ctx, m)
			}
		}, func(next MessageHandler) MessageHandler {
			return func(ctx context.Context, m *pubsub.Message) error {
				// this will overwrite the first interceptor
				m.Attributes = map[string]string{"intercepted": "true"}
				return next(ctx, m)
			}
		}),
	)

	wg := sync.WaitGroup{}
	var receivedCount int64 = 0
	err = subscriber.HandleSubscriptionFunc(sub.ID(), func(ctx context.Context, m *pubsub.Message) error {
		defer wg.Done()
		m.Ack()
		atomic.AddInt64(&receivedCount, 1)
		if m.Attributes["intercepted"] != "true" {
			t.Error("Interceptor didn't work")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	subscriber.Run(context.Background())

	var publishCount int64 = 2
	for i := 0; i < int(publishCount); i++ {
		wg.Add(1)
		ctx := context.Background()
		_, err := topic.Publish(ctx, &pubsub.Message{Data: []byte("test")}).Get(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}

	wg.Wait()

	if atomic.LoadInt64(&receivedCount) != publishCount {
		t.Errorf("Published count and received count doesn't match, published: %d, received: %d", publishCount, receivedCount)
	}
}

func TestSubscriber_Close(t *testing.T) {
	t.Parallel()

	pubsubClient, err := pubsub.NewClient(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}

	topic, err := pubsubClient.CreateTopic(context.Background(), fmt.Sprintf("TestSubscriber_Close_%d", time.Now().Unix()))
	if err != nil {
		t.Fatal(err)
	}

	sub, err := pubsubClient.CreateSubscription(
		context.Background(),
		fmt.Sprintf("TestSubscriber_Close_%d", time.Now().Unix()),
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err != nil {
		t.Fatal(err)
	}

	subscriber := NewSubscriber(pubsubClient)

	err = subscriber.HandleSubscriptionFunc(sub.ID(), func(ctx context.Context, m *pubsub.Message) error {
		m.Ack()
		t.Error("Must not received messages after close")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	subscriber.Run(context.Background())
	subscriber.Close()

	ctx := context.Background()
	if _, err := topic.Publish(ctx, &pubsub.Message{Data: []byte("test")}).Get(ctx); err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)
}
