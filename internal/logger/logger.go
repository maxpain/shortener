package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log = zap.NewNop()

func Init() {
	logger, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
	}

	Log = logger
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size

	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		startedAt := time.Now()
		next.ServeHTTP(&lw, r)
		duration := time.Since(startedAt)

		Log.Info("Incoming HTTP request",
			zap.String("Method", r.Method),
			zap.String("URI", r.URL.Path),
			zap.Duration("Duration", duration),
			zap.Int("Status", responseData.status),
			zap.Int("Size", responseData.size),
		)
	})
}
