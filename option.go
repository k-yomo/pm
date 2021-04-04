package pm

type Option interface {
	apply(*Config)
}

type optionFunc struct {
	f func(config *Config)
}

func (fdo *optionFunc) apply(do *Config) {
	fdo.f(do)
}

func newOptionFunc(f func(c *Config)) *optionFunc {
	return &optionFunc{
		f: f,
	}
}

func WithSubscriptionInterceptor(interceptors ...SubscriptionInterceptor) Option {
	return newOptionFunc(func(o *Config) {
		o.subscriptionInterceptors = interceptors
	})
}
