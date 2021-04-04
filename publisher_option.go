package pm

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

func WithPublishInterceptor(interceptors ...PublishInterceptor) PublisherOption {
	return newPublisherOptionFunc(func(o *PublisherConfig) {
		o.publishInterceptors = interceptors
	})
}
