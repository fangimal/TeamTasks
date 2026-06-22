package router

import (
	"log/slog"
	"net/http"

	httpdelivery "github.com/fangimal/TeamTask/internal/delivery/http"
	"github.com/fangimal/TeamTask/internal/delivery/http/middleware"
)

func New(logger *slog.Logger, healthChecker httpdelivery.HealthChecker) http.Handler {
	mux := http.NewServeMux()
	healthHandler := httpdelivery.NewHealthHandler(logger, healthChecker)

	mux.HandleFunc("GET /health", healthHandler.Check)

	return middleware.Logging(logger)(mux)
}
