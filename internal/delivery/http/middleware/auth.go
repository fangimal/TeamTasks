package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	jwtpkg "github.com/fangimal/TeamTask/pkg/jwt"
	"github.com/fangimal/TeamTask/pkg/response"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func Auth(logger *slog.Logger, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			token, ok := bearerToken(request.Header.Get("Authorization"))
			if !ok {
				response.WriteError(logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, err := jwtpkg.Validate(token, jwtSecret)
			if err != nil {
				response.WriteError(logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(request.Context(), userIDContextKey, claims.UserID)
			next.ServeHTTP(responseWriter, request.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", false
	}

	return strings.TrimSpace(token), true
}
