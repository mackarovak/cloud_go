package limiter

import (
	"sync"
	"time"
)

type ClientConfig struct {
	Capacity int `yaml:"capacity"`
	Rate     int `yaml:"rate"`
}

type Limiter struct {
	defaultConfig ClientConfig
	clientConfigs map[string]ClientConfig
	buckets       map[string]*TokenBucket
	mu            sync.RWMutex
	refillInterval time.Duration
	stopChan      chan struct{}
}

func NewLimiter(defaultCap, defaultRate int, refillInterval time.Duration, clientConfigs map[string]ClientConfig) *Limiter {
	l := &Limiter{
		defaultConfig: ClientConfig{defaultCap, defaultRate},
		clientConfigs: clientConfigs,
		buckets:       make(map[string]*TokenBucket),
		refillInterval: refillInterval,
		stopChan:      make(chan struct{}),
	}
	
	go l.backgroundRefill()
	return l
}

func (l *Limiter) Allow(clientID string) bool {
	l.mu.RLock()
	bucket, exists := l.buckets[clientID]
	l.mu.RUnlock()

	if !exists {
		config := l.getClientConfig(clientID)
		bucket = NewTokenBucket(config.Capacity, config.Rate, l.refillInterval)
		
		l.mu.Lock()
		l.buckets[clientID] = bucket
		l.mu.Unlock()
	}

	return bucket.Take()
}

func (l *Limiter) getClientConfig(clientID string) ClientConfig {
	if config, ok := l.clientConfigs[clientID]; ok {
		return config
	}
	return l.defaultConfig
}

func (l *Limiter) backgroundRefill() {
	ticker := time.NewTicker(l.refillInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			l.refillAllBuckets()
		case <-l.stopChan:
			return
		}
	}
}

func (l *Limiter) refillAllBuckets() {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	for _, bucket := range l.buckets {
		bucket.refill() 
	}
}

func (l *Limiter) Stop() {
	close(l.stopChan)
}