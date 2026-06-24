package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/fangimal/TeamTask/internal/config"
	"github.com/fangimal/TeamTask/pkg/response"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
	logger *slog.Logger
}

func NewRateLimiter(client *redis.Client, cfg config.RateLimitConfig, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		client: client,
		limit:  cfg.RequestsPerMinute,
		window: time.Minute,
		logger: logger,
	}
}

func (limiter *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if limiter.client == nil {
			next.ServeHTTP(responseWriter, request)
			return
		}

		key := limiter.resolveKey(request)

		count, err := limiter.client.Incr(request.Context(), key).Result()
		if err != nil {
			limiter.logger.ErrorContext(request.Context(), "rate limit redis error", slog.Any("error", err))
			next.ServeHTTP(responseWriter, request)
			return
		}

		if count == 1 {
			limiter.client.Expire(request.Context(), key, limiter.window)
		}

		if count > int64(limiter.limit) {
			responseWriter.Header().Set("Retry-After", strconv.Itoa(int(limiter.window.Seconds())))
			response.WriteError(limiter.logger, responseWriter, request, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(responseWriter, request)
	})
}

func (limiter *RateLimiter) resolveKey(request *http.Request) string {
	userID, ok := UserIDFromContext(request.Context())
	if ok {
		return fmt.Sprintf("rate_limit:%d", userID)
	}

	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		host = request.RemoteAddr
	}

	return fmt.Sprintf("rate_limit:ip:%s", host)
}
