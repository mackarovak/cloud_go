// internal/proxy/proxy.go
package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	
	"cloud/internal/limiter"
)

type ReverseProxy struct {
	proxy    *httputil.ReverseProxy
	limiter  *limiter.Limiter
}

func NewReverseProxy(target *url.URL, limiter *limiter.Limiter) *ReverseProxy {
	rp := &ReverseProxy{
		proxy:   httputil.NewSingleHostReverseProxy(target),
		limiter: limiter,
	}
	
	rp.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}
	
	return rp
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	if !rp.limiter.Allow(clientIP) {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}
	
	rp.proxy.ServeHTTP(w, r)
}

func getClientIP(r *http.Request) string {
	// Получаем реальный IP за прокси (если есть)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}