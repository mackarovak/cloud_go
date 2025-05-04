package tests

import (
	"net/http"
	"net/http/httptest"
	"cloud/internal/limiter"
	"cloud/internal/server"
	"testing"
	"time"
)

func BenchmarkRateLimiter(b *testing.B) {
    rl := limiter.NewLimiter(10000, 1000, time.Second, nil)
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    srv := server.NewServer(rl, handler)
    testServer := httptest.NewServer(srv)
    defer testServer.Close()
    
    client := &http.Client{
        Timeout: 3 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            IdleConnTimeout:     90 * time.Second,
            DisableKeepAlives:   false,
        },
    }
    defer client.CloseIdleConnections()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp, err := client.Get(testServer.URL)
            if err != nil {
                b.Errorf("Request failed: %v", err)
                continue
            }
            resp.Body.Close()
        }
    })
}

func BenchmarkBalancer(b *testing.B) {
    backends := []string{"http://backend1", "http://backend2", "http://backend3"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = backends[i%len(backends)]
        // Добавьте реальную логику балансировки вместо мока
    }
}
