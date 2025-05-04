package balancer

import (
	"net/http"
	"sync"
)

type Backend struct {
	URL     string
	Healthy bool
	mu      sync.RWMutex
}

func (b *Backend) SetHealthy(healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Healthy = healthy
}

func (b *Backend) IsHealthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Healthy
}

type Balancer interface {
	NextBackend() string
	UpdateBackends([]Backend)
	HealthCheck()
}

type RoundRobinBalancer struct {
	backends []Backend
	current  int
	mu       sync.Mutex
}

func NewRoundRobinBalancer(backends []Backend) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		backends: backends,
		current:  0,
	}
}

func (b *RoundRobinBalancer) NextBackend() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := 0; i < len(b.backends); i++ {
		backend := &b.backends[b.current]
		b.current = (b.current + 1) % len(b.backends)
		
		if backend.IsHealthy() {
			return backend.URL
		}
	}
	return ""
}

func (b *RoundRobinBalancer) UpdateBackends(backends []Backend) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.backends = backends
}

func (b *RoundRobinBalancer) HealthCheck() {
	for i := range b.backends {
		go func(backend *Backend) {
			resp, err := http.Get(backend.URL + "/health")
			backend.SetHealthy(err == nil && resp.StatusCode == http.StatusOK)
			if resp != nil {
				resp.Body.Close()
			}
		}(&b.backends[i])
	}
}