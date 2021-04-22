package pm_recovery

import (
	"context"
	"log"
)

type options struct {
	recoveryHandlerFunc RecoveryHandlerFunc
}

type Option func(*options)

func defaultRecoveryHandler(_ context.Context, p interface{}) {
	log.Printf("%v\n", p)
}

// WithRecoveryHandler customizes the function for recovering from a panic.
func WithRecoveryHandler(f RecoveryHandlerFunc) Option {
	return func(o *options) {
		o.recoveryHandlerFunc = f
	}
}

// WithDebugRecoveryHandler customizes the stdout of a panic to be more human readable.
func WithDebugRecoveryHandler() Option {
	return func(o *options) {
		o.recoveryHandlerFunc = func(ctx context.Context, p interface{}) {
			PrintPrettyStack(p)
		}
	}
}
