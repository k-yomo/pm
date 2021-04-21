package pm

// PublisherOption is a option to change publisher configuration.
type PublisherOption interface {
	apply(*publisherOptions)
}

type publisherOptionFunc struct {
	f func(*publisherOptions)
}

func (p *publisherOptionFunc) apply(po *publisherOptions) {
	p.f(po)
}

func newPublisherOptionFunc(f func(*publisherOptions)) *publisherOptionFunc {
	return &publisherOptionFunc{
		f: f,
	}
}

// WithPublishInterceptor sets publish interceptors.
func WithPublishInterceptor(interceptors ...PublishInterceptor) PublisherOption {
	return newPublisherOptionFunc(func(po *publisherOptions) {
		po.publishInterceptors = interceptors
	})
}
