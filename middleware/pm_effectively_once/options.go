package pm_effectively_once

const DefaultDeduplicateKey = "deduplicate_key"

type options struct {
	deduplicateKey string
}

type Option func(*options)

// WithCustomDeduplicateKey customizes the attribute key for de-duplicate key.
func WithCustomDeduplicateKey(key string) Option {
	return func(o *options) {
		o.deduplicateKey = key
	}
}
