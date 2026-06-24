package router

import (
	"log/slog"
	"net/http"

	"github.com/fangimal/TeamTask/internal/config"
	httpdelivery "github.com/fangimal/TeamTask/internal/delivery/http"
	"github.com/fangimal/TeamTask/internal/delivery/http/middleware"
	"github.com/fangimal/TeamTask/internal/usecase"
	"github.com/redis/go-redis/v9"
)

func New(
	logger *slog.Logger,
	dbChecker httpdelivery.HealthChecker,
	cacheChecker httpdelivery.HealthChecker,
	authUseCase httpdelivery.AuthUseCase,
	teamUseCase *usecase.TeamUseCase,
	taskUseCase *usecase.TaskUseCase,
	commentUseCase *usecase.CommentUseCase,
	analyticsUseCase *usecase.AnalyticsUseCase,
	jwtSecret string,
	redisClient *redis.Client,
	rateLimitCfg config.RateLimitConfig,
) http.Handler {
	mux := http.NewServeMux()
	healthHandler := httpdelivery.NewHealthHandler(logger, dbChecker, cacheChecker)
	authHandler := httpdelivery.NewAuthHandler(logger, authUseCase)
	protectedHandler := httpdelivery.NewProtectedHandler(logger)
	teamHandler := httpdelivery.NewTeamHandler(logger, teamUseCase)
	taskHandler := httpdelivery.NewTaskHandler(logger, taskUseCase)
	commentHandler := httpdelivery.NewCommentHandler(logger, commentUseCase)
	analyticsHandler := httpdelivery.NewAnalyticsHandler(logger, analyticsUseCase)

	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

	authMiddleware := middleware.Auth(logger, jwtSecret)
	rateLimiter := middleware.NewRateLimiter(redisClient, rateLimitCfg, logger)

	protect := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware(rateLimiter.Middleware(http.HandlerFunc(handler)))
	}

	mux.Handle("GET /api/v1/ping", protect(protectedHandler.Ping))
	mux.Handle("POST /api/v1/teams", protect(teamHandler.CreateTeam))
	mux.Handle("GET /api/v1/teams", protect(teamHandler.GetUserTeams))
	mux.Handle("POST /api/v1/teams/{id}/invite", protect(teamHandler.InviteUser))

	mux.Handle("POST /api/v1/tasks", protect(taskHandler.CreateTask))
	mux.Handle("GET /api/v1/tasks", protect(taskHandler.GetTasks))
	mux.Handle("GET /api/v1/tasks/{id}", protect(taskHandler.GetTaskByID))
	mux.Handle("PUT /api/v1/tasks/{id}", protect(taskHandler.UpdateTask))
	mux.Handle("GET /api/v1/tasks/{id}/history", protect(taskHandler.GetTaskHistory))

	mux.Handle("POST /api/v1/tasks/{id}/comments", protect(commentHandler.CreateComment))
	mux.Handle("GET /api/v1/tasks/{id}/comments", protect(commentHandler.GetComments))

	mux.Handle("GET /api/v1/analytics/team-stats", protect(analyticsHandler.GetTeamStats))
	mux.Handle("GET /api/v1/analytics/top-users", protect(analyticsHandler.GetTopUsers))
	mux.Handle("GET /api/v1/analytics/integrity-check", protect(analyticsHandler.GetIntegrityCheck))

	return middleware.Logging(logger)(mux)
}
