package tests

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	
	"cloud/internal/limiter"
	"cloud/internal/server"
)

func TestRateLimiting(t *testing.T) {
	// 1. Инициализируем rate limiter
	rl := limiter.NewLimiter(
		5,              // default capacity
		1,              // default rate
		time.Second,    // refill interval
		make(map[string]limiter.ClientConfig),
	)

	// 2. Создаем тестовый HTTP-обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 3. Создаем тестовый сервер
	srv := server.NewServer(rl, handler)
	testServer := httptest.NewServer(srv)
	defer testServer.Close()

	t.Run("Basic rate limiting", func(t *testing.T) {
		client := testServer.Client()

		// Первые 5 запросов должны проходить
		for i := 0; i < 5; i++ {
			resp, err := client.Get(testServer.URL)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		}

		// 6-й запрос должен быть отклонен
		resp, err := client.Get(testServer.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected status 429, got %d", resp.StatusCode)
		}
	})

	t.Run("Parallel requests", func(t *testing.T) {
		var wg sync.WaitGroup
		successful := 0
		rateLimited := 0
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := testServer.Client()
				resp, err := client.Get(testServer.URL)
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				switch resp.StatusCode {
				case http.StatusOK:
					successful++
				case http.StatusTooManyRequests:
					rateLimited++
				default:
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			if err != nil {
				t.Error(err)
			}
		}

		t.Logf("Successful: %d, Rate limited: %d", successful, rateLimited)
	})
}