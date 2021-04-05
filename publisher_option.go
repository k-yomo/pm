package pm

// PublisherOption is a option to change publisher configuration.
type PublisherOption interface {
	apply(*publisherOptions)
}

type publisherOptionFunc struct {
	f func(config *publisherOptions)
}

func (p *publisherOptionFunc) apply(pc *publisherOptions) {
	p.f(pc)
}

func newPublisherOptionFunc(f func(pc *publisherOptions)) *publisherOptionFunc {
	return &publisherOptionFunc{
		f: f,
	}
}

// WithPublishInterceptor sets publish interceptors.
func WithPublishInterceptor(interceptors ...PublishInterceptor) PublisherOption {
	return newPublisherOptionFunc(func(sc *publisherOptions) {
		sc.publishInterceptors = interceptors
	})
}
