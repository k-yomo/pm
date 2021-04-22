package pm_recovery

import (
	"context"
	"log"
)

type options struct {
	recoveryHandlerFunc RecoveryHandlerFunc
}

type Option func(*options)

func defaultRecoveryHandlerContext(_ context.Context, p interface{}) {
	log.Printf("%v\n", p)
}

// RecoveryHandlerFunc customizes the function for recovering from a panic.
func WithRecoveryHandlerContext(f RecoveryHandlerFunc) Option {
	return func(o *options) {
		o.recoveryHandlerFunc = f
	}
}

// WithDebugMode customizes the stdout of a panic to be more human readable.
func WithDebugMode() Option {
	return func(o *options) {
		o.recoveryHandlerFunc = func(ctx context.Context, p interface{}) {
			PrintPrettyStack(p)
		}
	}
}
