package router

import (
	"log/slog"
	"net/http"

	httpdelivery "github.com/fangimal/TeamTask/internal/delivery/http"
	"github.com/fangimal/TeamTask/internal/delivery/http/middleware"
)

func New(
	logger *slog.Logger,
	healthChecker httpdelivery.HealthChecker,
	authUseCase httpdelivery.AuthUseCase,
	jwtSecret string,
) http.Handler {
	mux := http.NewServeMux()
	healthHandler := httpdelivery.NewHealthHandler(logger, healthChecker)
	authHandler := httpdelivery.NewAuthHandler(logger, authUseCase)
	protectedHandler := httpdelivery.NewProtectedHandler(logger)

	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)
	mux.Handle("GET /api/v1/ping", middleware.Auth(logger, jwtSecret)(http.HandlerFunc(protectedHandler.Ping)))

	return middleware.Logging(logger)(mux)
}
