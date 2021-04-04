package pm

// SubscriberOption is a option to change subscriber configuration.
type SubscriberOption interface {
	apply(*SubscriberConfig)
}

type subscriberOptionFunc struct {
	f func(config *SubscriberConfig)
}

func (fdo *subscriberOptionFunc) apply(do *SubscriberConfig) {
	fdo.f(do)
}

func newSubscriberOptionFunc(f func(c *SubscriberConfig)) *subscriberOptionFunc {
	return &subscriberOptionFunc{
		f: f,
	}
}

// WithSubscriptionInterceptor sets subscription interceptors.
func WithSubscriptionInterceptor(interceptors ...SubscriptionInterceptor) SubscriberOption {
	return newSubscriberOptionFunc(func(o *SubscriberConfig) {
		o.subscriptionInterceptors = interceptors
	})
}
