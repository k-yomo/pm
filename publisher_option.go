package pm

// PublisherOption is a option to change publisher configuration.
type PublisherOption interface {
	apply(*PublisherConfig)
}

type publisherOptionFunc struct {
	f func(config *PublisherConfig)
}

func (fdo *publisherOptionFunc) apply(do *PublisherConfig) {
	fdo.f(do)
}

func newPublisherOptionFunc(f func(c *PublisherConfig)) *publisherOptionFunc {
	return &publisherOptionFunc{
		f: f,
	}
}

// WithPublishInterceptor sets publish interceptors.
func WithPublishInterceptor(interceptors ...PublishInterceptor) PublisherOption {
	return newPublisherOptionFunc(func(o *PublisherConfig) {
		o.publishInterceptors = interceptors
	})
}
