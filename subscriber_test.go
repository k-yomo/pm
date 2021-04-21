package pm

import (
	"cloud.google.com/go/pubsub"
	"reflect"
	"testing"
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
