package logger

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// HTTPMiddleware creates a logging middleware for HTTP requests
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // default
		}

		// Log incoming request
		Get().Debug("incoming request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Log request completion
		duration := time.Since(start)

		logFunc := Get().Debug
		if wrapped.statusCode >= 500 {
			logFunc = Get().Error
		} else if wrapped.statusCode >= 400 {
			logFunc = Get().Warn
		}

		logFunc("request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.statusCode),
			slog.Duration("duration", duration),
			slog.Int64("bytes", wrapped.written),
		)
	})
}

// HTTPMiddlewareFunc is a function version that wraps http.HandlerFunc
func HTTPMiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HTTPMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}
