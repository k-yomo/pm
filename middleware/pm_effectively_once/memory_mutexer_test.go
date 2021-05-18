package pm_effectively_once

import (
	"context"
	"errors"
	"testing"
)

func Test_memoryMutexer_RunInTx(t *testing.T) {
	t.Parallel()

	t.Run("an event with already processed id is not processed", func(t *testing.T) {
		t.Parallel()

		mutexer := NewMemoryMutexer()
		err := mutexer.RunInTx(context.Background(), "test", func() error {
			return nil
		})
		if err != nil {
			t.Errorf("memoryMutexer.RunInTx is expected to return nil, but got err: %v", err)
		}

		var processed bool
		err = mutexer.RunInTx(context.Background(), "test", func() error {
			processed = true
			return nil
		})
		if err != nil {
			t.Errorf("memoryMutexer.RunInTx is expected to return nil, but got err: %v", err)
		}
		if processed {
			t.Errorf("memoryMutexer.RunInTx must discard an event with the already processed de-duplicate key")
		}
	})

	t.Run("when processing first event returns error, next event with same de-duplicate key is processed", func(t *testing.T) {
		t.Parallel()

		mutexer := NewMemoryMutexer()
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
			t.Errorf("memoryMutexer.RunInTx must process an event with not processed de-duplicate key")
		}
	})
}
