package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/Dokhoyan/marketplace/internal/logger"
	"go.uber.org/zap"
)

type contextKey string

const (
	requestIDContextKey contextKey = "requestID"
	requestIDHeader     string     = "X-Request-ID"
	requestIDField      string     = "request_id"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000")
	}
	return hex.EncodeToString(b)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		w.Header().Set(requestIDHeader, requestID)

		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		r = r.WithContext(ctx)

		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path = path + "?" + r.URL.RawQuery
		}

		logger.Info("request started",
			zap.String(requestIDField, requestID),
			zap.String("method", r.Method),
			zap.String("path", path),
			zap.String("ip", r.RemoteAddr),
			zap.String("user-agent", r.UserAgent()),
		)

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		end := time.Now().UTC()
		latency := end.Sub(start)

		fields := []zap.Field{
			zap.String(requestIDField, requestID),
			zap.Int("status", rw.statusCode),
			zap.String("method", r.Method),
			zap.String("path", path),
			zap.String("query", r.URL.RawQuery),
			zap.String("ip", r.RemoteAddr),
			zap.String("user-agent", r.UserAgent()),
			zap.Duration("latency", latency),
		}

		status := rw.statusCode
		switch {
		case status >= 500:
			logger.Error("request completed", fields...)
		case status >= 400:
			logger.Warn("request completed", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	})
}

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if requestID, ok := ctx.Value(requestIDContextKey).(string); ok {
		return requestID
	}

	return ""
}
