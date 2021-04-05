package pm

// SubscriberOption is a option to change subscriber configuration.
type SubscriberOption interface {
	apply(*SubscriberConfig)
}

type subscriberOptionFunc struct {
	f func(config *SubscriberConfig)
}

func (s *subscriberOptionFunc) apply(sc *SubscriberConfig) {
	s.f(sc)
}

func newSubscriberOptionFunc(f func(sc *SubscriberConfig)) *subscriberOptionFunc {
	return &subscriberOptionFunc{
		f: f,
	}
}

// WithSubscriptionInterceptor sets subscription interceptors.
func WithSubscriptionInterceptor(interceptors ...SubscriptionInterceptor) SubscriberOption {
	return newSubscriberOptionFunc(func(sc *SubscriberConfig) {
		sc.subscriptionInterceptors = interceptors
	})
}
