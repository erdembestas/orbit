package httpserver

import (
	"net/http"
	"time"

	"orbit/internal/config"
)

type Logger interface {
	Info(msg string, attrs ...any)
	Error(msg string, attrs ...any)
}

func New(cfg config.Config, readyChecker ReadyChecker, authService AuthService, dataStore DataStore, evidenceService EvidenceService, reasoner Reasoner, ready *uint32, logger Logger) *http.Server {
	handler := withLogging(newHandler(cfg, readyChecker, authService, dataStore, evidenceService, reasoner, ready), logger)

	return &http.Server{
		Addr:              cfg.Address(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func withLogging(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		logger.Info(
			"request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
