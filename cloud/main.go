package main

import (
	"log"
	"net/http"
	"strconv"
	"net/url"
	"cloud/internal/balancer"
	"cloud/internal/config"
	"cloud/internal/limiter"
	"cloud/internal/proxy"
	"cloud/internal/server"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// В main.go
clientConfigs := make(map[string]limiter.ClientConfig)
for k, v := range cfg.RateLimiter.ClientSpecific {
	clientConfigs[k] = limiter.ClientConfig{
		Capacity: v.Capacity,
		Rate:     v.Rate,
	}
}

rl := limiter.NewLimiter(
	cfg.RateLimiter.DefaultCapacity,
	cfg.RateLimiter.DefaultRate,
	cfg.RateLimiter.RefillInterval,
	clientConfigs,
)
	
	// Создаем бэкенды для балансировщика
	var backends []balancer.Backend
	for _, backend := range cfg.Backends {
		backends = append(backends, balancer.Backend{
			URL:     backend.URL,
			Healthy: backend.Healthy,
		})
	}

	// Инициализация балансировщика
	b := balancer.NewRoundRobinBalancer(backends)

	// Создание прокси с rate limiting
	proxyHandler := func(w http.ResponseWriter, r *http.Request) {
		backendURL := b.NextBackend()
		if backendURL == "" {
			http.Error(w, "No available backends", http.StatusServiceUnavailable)
			return
		}

		target, err := url.Parse(backendURL)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		proxy.NewReverseProxy(target, rl).ServeHTTP(w, r)
	}

	// Запуск сервера
	srv := server.NewServer(rl, http.HandlerFunc(proxyHandler))
	log.Printf("Starting server on :%d with rate limiting", cfg.Port)
	log.Fatal(srv.Start(":" + strconv.Itoa(cfg.Port)))
}