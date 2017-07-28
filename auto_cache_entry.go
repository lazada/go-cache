package cache

import (
	"sync"
	"time"

	"go-cache/errors"
)

// EntryAutoCache contains data
type EntryAutoCache struct {
	name     string
	value    interface{}
	updater  func() (interface{}, error)
	interval time.Duration
	run      bool
	ticker   *time.Ticker
	mutex    sync.RWMutex
	signals  chan struct{}
	logger   IAutoCacheLogger
}

// CreateEntryAutoCache returns new instance of EntryAutoCache
func CreateEntryAutoCache(updater func() (interface{}, error), interval time.Duration, name string, logger IAutoCacheLogger) *EntryAutoCache {
	return &EntryAutoCache{
		name:     name,
		run:      false,
		updater:  updater,
		interval: interval,
		logger:   logger,
	}
}

// GetValue returns the result of processing the updater
func (entry *EntryAutoCache) GetValue() (interface{}, error) {
	entry.mutex.RLock()

	if !entry.run {
		entry.mutex.RUnlock()
		if err := entry.process(); err != nil {
			return nil, err
		}

		entry.mutex.RLock()
	}

	defer entry.mutex.RUnlock()

	if entry.value == nil {
		return nil, errors.Errorf("Value is not set")
	}

	return entry.value, nil
}

// Start starts updater process in goroutine
func (entry *EntryAutoCache) Start() error {
	if entry.run {
		return nil
	}
	entry.run = true
	entry.signals = make(chan struct{}, 1)
	entry.ticker = time.NewTicker(entry.interval)
	if err := entry.process(); err != nil {
		return err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				entry.logger.Criticalf("Panic in entry.loop(), %v", r)
			}
		}()
		entry.loop()
	}()
	return nil
}

// Stop stops updater process
func (entry *EntryAutoCache) Stop() {
	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	if entry.run {
		entry.run = false
		entry.value = nil
		entry.ticker.Stop()

		entry.signals <- struct{}{}
	}
}

func (entry *EntryAutoCache) process() error {
	entry.mutex.RLock()
	updater := entry.updater
	name := entry.name
	entry.mutex.RUnlock()

	if value, err := updater(); err == nil {
		entry.mutex.Lock()
		entry.value = value
		entry.mutex.Unlock()
	} else {
		entry.logger.Errorf("Auto cache updater \"%s\" error: %s", name, err)
		return err
	}

	return nil
}

func (entry *EntryAutoCache) loop() {
	for {
		select {
		case <-entry.ticker.C:
			entry.process()
		case <-entry.signals:
			return
		}
	}
}
