package pm_effectively_once

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func Test_redisMutexer_RunInTx(t *testing.T) {
	t.Parallel()

	redisClient := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_URL")})

	t.Run("an event with already processed id is not processed", func(t *testing.T) {
		t.Parallel()

		mutexer := NewRedisMutexer(redisClient, randString(t, 20), 1*time.Hour)
		err := mutexer.RunInTx(context.Background(), "test", func() error {
			return nil
		})
		if err != nil {
			t.Errorf("redisMutexer.RunInTx is expected to return nil, but got err: %v", err)
		}

		var processed bool
		err = mutexer.RunInTx(context.Background(), "test", func() error {
			processed = true
			return nil
		})
		if err != nil {
			t.Errorf("redisMutexer.RunInTx is expected to return nil, but got err: %v", err)
		}
		if processed {
			t.Errorf("redisMutexer.RunInTx must discard an event with the already processed id")
		}
	})

	t.Run("when processing first event returns error, next event with same id is processed", func(t *testing.T) {
		t.Parallel()

		mutexer := NewRedisMutexer(redisClient, randString(t, 20), 1*time.Hour)
		err := mutexer.RunInTx(context.Background(), "test", func() error {
			return errors.New("test")
		})
		if err == nil {
			t.Error("memoryMutexer.RunInTx is expected to return err, but got nil")
		}

		var processed bool
		err = mutexer.RunInTx(context.Background(), "test", func() error {
			processed = true
			return nil
		})
		if err != nil {
			t.Errorf("memoryMutexer.RunInTx is expected to return nil, but got err: %v", err)
		}
		if !processed {
			t.Errorf("memoryMutexer.RunInTx must process an event with not processed id")
		}
	})
}
