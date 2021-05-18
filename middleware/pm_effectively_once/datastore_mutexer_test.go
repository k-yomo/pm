package pm_effectively_once

import (
	"cloud.google.com/go/datastore"
	"context"
	"errors"
	"testing"
)

func Test_datastoreMutexer_RunInTx(t *testing.T) {
	t.Parallel()

	dsClient, err := datastore.NewClient(context.Background(), "test")
	if err != nil {
		t.Fatalf("initialize datastore client failed: %v", err)
	}

	t.Run("an event with already processed id is not processed", func(t *testing.T) {
		t.Parallel()

		mutexer := NewDatastoreMutexer(randString(t, 20), dsClient)
		err := mutexer.RunInTx(context.Background(), "test", func() error {
			return nil
		})
		if err != nil {
			t.Errorf("datastoreMutexer.RunInTx is expected to return nil, but got err: %v", err)
		}

		var processed bool
		err = mutexer.RunInTx(context.Background(), "test", func() error {
			processed = true
			return nil
		})
		if err != nil {
			t.Errorf("datastoreMutexer.RunInTx is expected to return nil, but got err: %v", err)
		}
		if processed {
			t.Errorf("datastoreMutexer.RunInTx must discard an event with the already processed de-duplicate key")
		}
	})

	t.Run("when processing first event returns error, next event with same de-duplicate key is processed", func(t *testing.T) {
		t.Parallel()

		mutexer := NewDatastoreMutexer(randString(t, 20), dsClient)
		err := mutexer.RunInTx(context.Background(), "test", func() error {
			return errors.New("test")
		})
		if err == nil {
			t.Error("datastoreMutexer.RunInTx is expected to return err, but got nil")
		}

		var processed bool
		err = mutexer.RunInTx(context.Background(), "test", func() error {
			processed = true
			return nil
		})
		if err != nil {
			t.Errorf("datastoreMutexer.RunInTx is expected to return nil, but got err: %v", err)
		}
		if !processed {
			t.Errorf("datastoreMutexer.RunInTx must process an event with not processed id")
		}
	})
}
