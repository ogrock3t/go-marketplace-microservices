package logger

import (
	"log/slog"
	"net/http"
	"time"
)

func RequestLogger(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			log.Info("request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("duration", time.Since(start).String()),
			)
		})
	}
}
