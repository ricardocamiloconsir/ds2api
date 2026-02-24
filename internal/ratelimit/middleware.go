package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type Middleware struct {
	perIPPerMinute  int
	globalPerMinute int
	mu              sync.Mutex
	windowStart     time.Time
	globalCount     int
	perIPCount      map[string]int
}

func New(perIPPerMinute, globalPerMinute, _ int) *Middleware {
	return &Middleware{
		perIPPerMinute:  perIPPerMinute,
		globalPerMinute: globalPerMinute,
		windowStart:     time.Now(),
		perIPCount:      map[string]int{},
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m == nil || (m.perIPPerMinute <= 0 && m.globalPerMinute <= 0) {
			next.ServeHTTP(w, r)
			return
		}
		if !m.allow(clientIP(r.RemoteAddr)) {
			writeRateLimit(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) allow(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if now.Sub(m.windowStart) >= time.Minute {
		m.windowStart = now
		m.globalCount = 0
		m.perIPCount = map[string]int{}
	}
	if m.globalPerMinute > 0 && m.globalCount >= m.globalPerMinute {
		return false
	}
	if m.perIPPerMinute > 0 && m.perIPCount[ip] >= m.perIPPerMinute {
		return false
	}
	m.globalCount++
	m.perIPCount[ip]++
	return true
}

func writeRateLimit(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(`{"error":{"type":"rate_limit_error","message":"Too Many Requests"}}`))
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
