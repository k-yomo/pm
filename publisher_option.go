package pm

// PublisherOption is a option to change publisher configuration.
type PublisherOption interface {
	apply(*PublisherConfig)
}

type publisherOptionFunc struct {
	f func(config *PublisherConfig)
}

func (p *publisherOptionFunc) apply(pc *PublisherConfig) {
	p.f(pc)
}

func newPublisherOptionFunc(f func(pc *PublisherConfig)) *publisherOptionFunc {
	return &publisherOptionFunc{
		f: f,
	}
}

// WithPublishInterceptor sets publish interceptors.
func WithPublishInterceptor(interceptors ...PublishInterceptor) PublisherOption {
	return newPublisherOptionFunc(func(sc *PublisherConfig) {
		sc.publishInterceptors = interceptors
	})
}
