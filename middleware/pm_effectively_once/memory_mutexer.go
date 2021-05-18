package pm_effectively_once

import (
	"context"
	"sync"
)

type memoryMutexer struct {
	sync.Mutex
	processedKeyMap map[string]bool
}

func NewMemoryMutexer() Mutexer {
	return &memoryMutexer{processedKeyMap: make(map[string]bool)}
}

func (d *memoryMutexer) RunInTx(_ context.Context, deduplicateKey string, f func() error) error {
	d.Lock()
	defer d.Unlock()

	if d.processedKeyMap[deduplicateKey] {
		// the event already processed
		return nil
	}

	if err := f(); err != nil {
		return err
	}

	d.processedKeyMap[deduplicateKey] = true
	return nil
}
