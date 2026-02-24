package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"ds2api/internal/config"
)

type requestMetrics struct {
	startedAt      time.Time
	requestsTotal  atomic.Uint64
	errorsTotal    atomic.Uint64
	latencyTotalNs atomic.Uint64
	byStatus       sync.Map
}

type statusCounter struct {
	value atomic.Uint64
}

func newRequestMetrics() *requestMetrics {
	return &requestMetrics{startedAt: time.Now()}
}

func (m *requestMetrics) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		elapsed := time.Since(start)
		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}
		m.requestsTotal.Add(1)
		m.latencyTotalNs.Add(uint64(elapsed.Nanoseconds()))
		if status >= 500 {
			m.errorsTotal.Add(1)
		}
		counterAny, _ := m.byStatus.LoadOrStore(status, &statusCounter{})
		counterAny.(*statusCounter).value.Add(1)
	})
}

func (m *requestMetrics) handleMetrics(w http.ResponseWriter, _ *http.Request, poolStatus map[string]any) {
	total := m.requestsTotal.Load()
	avgMs := float64(0)
	if total > 0 {
		avgMs = float64(m.latencyTotalNs.Load()) / float64(total) / float64(time.Millisecond)
	}
	statuses := map[int]uint64{}
	m.byStatus.Range(func(key, value any) bool {
		k, ok1 := key.(int)
		v, ok2 := value.(*statusCounter)
		if ok1 && ok2 {
			statuses[k] = v.value.Load()
		}
		return true
	})
	payload := map[string]any{
		"service":        "ds2api",
		"uptime_seconds": int(time.Since(m.startedAt).Seconds()),
		"requests": map[string]any{
			"total":               total,
			"errors_5xx":          m.errorsTotal.Load(),
			"avg_duration_ms":     avgMs,
			"responses_by_status": statuses,
		},
		"pool": poolStatus,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}
		config.Logger.Info("[http] request",
			"request_id", middleware.GetReqID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", ww.BytesWritten(),
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_ip", r.RemoteAddr,
		)
	})
}
