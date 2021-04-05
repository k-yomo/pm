package pm

// SubscriberOption is a option to change subscriber configuration.
type SubscriberOption interface {
	apply(*subscriberOptions)
}

type subscriberOptionFunc struct {
	f func(config *subscriberOptions)
}

func (s *subscriberOptionFunc) apply(sc *subscriberOptions) {
	s.f(sc)
}

func newSubscriberOptionFunc(f func(sc *subscriberOptions)) *subscriberOptionFunc {
	return &subscriberOptionFunc{
		f: f,
	}
}

// WithSubscriptionInterceptor sets subscription interceptors.
func WithSubscriptionInterceptor(interceptors ...SubscriptionInterceptor) SubscriberOption {
	return newSubscriberOptionFunc(func(sc *subscriberOptions) {
		sc.subscriptionInterceptors = interceptors
	})
}
