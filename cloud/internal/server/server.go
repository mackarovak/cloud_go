package server

import (
	"net/http"
	"cloud/internal/limiter"
	"github.com/gorilla/mux"
)

type Server struct {
	limiter *limiter.Limiter
	router  *mux.Router
}

func NewServer(limiter *limiter.Limiter, handler http.Handler) *Server {
	s := &Server{
		limiter: limiter,
		router:  mux.NewRouter(),
	}
	
	// Настраиваем маршруты и middleware
	s.setupRoutes(handler)
	
	return s
}

func (s *Server) setupRoutes(handler http.Handler) {
	// Применяем middleware
	s.router.Use(s.rateLimitMiddleware)
	
	// Устанавливаем обработчик
	s.router.Handle("/", handler)
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.RemoteAddr
		if !s.limiter.Allow(clientID) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Реализуем интерфейс http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}