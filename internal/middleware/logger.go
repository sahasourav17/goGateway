package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func NewStructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			latency := time.Since(start)
			logger.Info("request completed",
				"status", ww.Status(),
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"request_id", middleware.GetReqID(r.Context()),
				"latency_ms", float64(latency.Nanoseconds())/1000000.0,
			)
		})
	}
}

func InitLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
