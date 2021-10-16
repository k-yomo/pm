package pm_effectively_once

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisMutexer struct {
	redisClient  *redis.Client
	keyPrefix    string
	lockDuration time.Duration
}

func NewRedisMutexer(redisClient *redis.Client, keyPrefix string, lockDuration time.Duration) Mutexer {
	return &redisMutexer{
		redisClient:  redisClient,
		keyPrefix:    keyPrefix,
		lockDuration: lockDuration,
	}
}

func (d *redisMutexer) RunInTx(ctx context.Context, deduplicateKey string, f func() error) error {
	key := d.keyPrefix + ":" + deduplicateKey
	return d.redisClient.Watch(ctx, func(tx *redis.Tx) error {
		if err := tx.Get(ctx, key).Err(); err == nil {
			// the event already processed
			return nil
		} else {
			if err != redis.Nil {
				return err
			}
		}
		_, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			if err := p.Set(ctx, key, true, d.lockDuration).Err(); err != nil {
				return err
			}

			return f()
		})
		return err
	})
}
