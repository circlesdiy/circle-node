package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call next handler
			next.ServeHTTP(rw, r)

			// Calculate request duration
			duration := time.Since(start)

			// Build log fields
			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", rw.statusCode),
				zap.Duration("duration", duration),
				zap.Int("bytes", rw.bytesWritten),
				zap.String("ip", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			}

			// Add HTMX-specific fields if present
			if IsHTMXRequest(r) {
				fields = append(fields, zap.Bool("htmx", true))
				if target := GetHTMXTarget(r); target != "" {
					fields = append(fields, zap.String("htmx_target", target))
				}
			}

			// Log with appropriate level based on status code
			switch {
			case rw.statusCode >= 500:
				logger.Error("http request", fields...)
			case rw.statusCode >= 400:
				logger.Warn("http request", fields...)
			default:
				logger.Info("http request", fields...)
			}
		})
	}
}
