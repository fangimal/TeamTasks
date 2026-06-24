package monitoring

import (
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (writer *statusWriter) WriteHeader(code int) {
	writer.statusCode = code
	writer.ResponseWriter.WriteHeader(code)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		recorder := &statusWriter{
			ResponseWriter: response,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, request)

		status := recorder.statusCode
		path := request.URL.Path

		HTTPRequestsTotal.WithLabelValues(request.Method, path, http.StatusText(status)).Inc()
		HTTPRequestDurationSeconds.WithLabelValues(request.Method, path, http.StatusText(status)).Observe(time.Since(startedAt).Seconds())
	})
}
