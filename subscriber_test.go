package pm

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
)

type subscriberOptionForTest struct {
}

func (s subscriberOptionForTest) apply(so *subscriberOptions) {
	so.subscriptionInterceptors = []SubscriptionInterceptor{nil}
}

func TestNewSubscriber(t *testing.T) {
	t.Parallel()

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
				subscriptionHandlers: map[string]*subscriptionHandler{},
				cancel:               nil,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

	topic, err := pubsubClient.CreateTopic(context.Background(), fmt.Sprintf("TestSubscriber_HandleSubscriptionFunc_%d", time.Now().Unix()))
	if err != nil {
		t.Fatal(err)
	}

	sub, err := pubsubClient.CreateSubscription(
		context.Background(),
		fmt.Sprintf("TestSubscriber_HandleSubscriptionFunc_%d", time.Now().Unix()),
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		subscription *pubsub.Subscription
		f            MessageHandler
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
				subscription: sub,
				f:            func(ctx context.Context, m *pubsub.Message) error { return nil },
			},
		},
		{
			name: "when a handler is already registered for the give subscription id, it returns error",
			subscriber: func() *Subscriber {
				s := NewSubscriber(pubsubClient)
				_ = s.HandleSubscriptionFunc(sub, func(ctx context.Context, m *pubsub.Message) error { return nil })
				return s
			}(),
			args: args{
				subscription: sub,
				f:            func(ctx context.Context, m *pubsub.Message) error { return nil },
			},
			wantErr: true,
		},
		{
			name:       "when the subscription does not exist, it returns error",
			subscriber: NewSubscriber(pubsubClient),
			args: args{
				subscription: pubsubClient.Subscription("invalid"),
				f:            nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.subscriber.HandleSubscriptionFunc(tt.args.subscription, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("HandleSubscriptionFunc() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if _, ok := tt.subscriber.subscriptionHandlers[tt.args.subscription.ID()]; !ok {
					t.Errorf("HandleSubscriptionFunc() is expected to set subscription handler")
				}
			}
		})
	}
}

func TestSubscriber_HandleSubscriptionFuncMap(t *testing.T) {
	t.Parallel()

	pubsubClient, err := pubsub.NewClient(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}

	topic, err := pubsubClient.CreateTopic(context.Background(), fmt.Sprintf("TestSubscriber_HandleSubscriptionFuncMap_%d", time.Now().Unix()))
	if err != nil {
		t.Fatal(err)
	}

	sub, err := pubsubClient.CreateSubscription(
		context.Background(),
		fmt.Sprintf("TestSubscriber_HandleSubscriptionFuncMap_%d", time.Now().Unix()),
		pubsub.SubscriptionConfig{Topic: topic},
	)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		funcMap map[*pubsub.Subscription]MessageHandler
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
				funcMap: map[*pubsub.Subscription]MessageHandler{
					sub: func(ctx context.Context, m *pubsub.Message) error { return nil },
				},
			},
		},
		{
			name:       "when the subscription does not exist, it returns error",
			subscriber: NewSubscriber(pubsubClient),
			args: args{
				funcMap: map[*pubsub.Subscription]MessageHandler{
					pubsubClient.Subscription("invalid"): func(ctx context.Context, m *pubsub.Message) error { return nil },
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.subscriber.HandleSubscriptionFuncMap(tt.args.funcMap); (err != nil) != tt.wantErr {
				t.Errorf("HandleSubscriptionFuncMap() error = %v, wantErr %v", err, tt.wantErr)
			}

			for sub, _ := range tt.args.funcMap {
				if _, ok := tt.subscriber.subscriptionHandlers[sub.ID()]; ok == tt.wantErr {
					t.Errorf("HandleSubscriptionFuncMap() set: %v, want set: %v", ok, !tt.wantErr)
				}
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
		WithSubscriptionInterceptor(func(_ *SubscriptionInfo, next MessageHandler) MessageHandler {
			return func(ctx context.Context, m *pubsub.Message) error {
				m.Attributes = map[string]string{"intercepted": "abc"}
				return next(ctx, m)
			}
		}, func(_ *SubscriptionInfo, next MessageHandler) MessageHandler {
			return func(ctx context.Context, m *pubsub.Message) error {
				// this will overwrite the first interceptor
				m.Attributes = map[string]string{"intercepted": "true"}
				return next(ctx, m)
			}
		}),
	)

	wg := sync.WaitGroup{}
	var receivedCount int64 = 0
	err = subscriber.HandleSubscriptionFunc(sub, func(ctx context.Context, m *pubsub.Message) error {
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

	err = subscriber.HandleSubscriptionFunc(sub, func(ctx context.Context, m *pubsub.Message) error {
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
