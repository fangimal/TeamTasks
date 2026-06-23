package router

import (
	"log/slog"
	"net/http"

	httpdelivery "github.com/fangimal/TeamTask/internal/delivery/http"
	"github.com/fangimal/TeamTask/internal/delivery/http/middleware"
	"github.com/fangimal/TeamTask/internal/usecase"
)

func New(
	logger *slog.Logger,
	healthChecker httpdelivery.HealthChecker,
	authUseCase httpdelivery.AuthUseCase,
	teamUseCase *usecase.TeamUseCase,
	taskUseCase *usecase.TaskUseCase,
	commentUseCase *usecase.CommentUseCase,
	jwtSecret string,
) http.Handler {
	mux := http.NewServeMux()
	healthHandler := httpdelivery.NewHealthHandler(logger, healthChecker)
	authHandler := httpdelivery.NewAuthHandler(logger, authUseCase)
	protectedHandler := httpdelivery.NewProtectedHandler(logger)
	teamHandler := httpdelivery.NewTeamHandler(logger, teamUseCase)
	taskHandler := httpdelivery.NewTaskHandler(logger, taskUseCase)
	commentHandler := httpdelivery.NewCommentHandler(logger, commentUseCase)

	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)
	mux.Handle("GET /api/v1/ping", middleware.Auth(logger, jwtSecret)(http.HandlerFunc(protectedHandler.Ping)))

	authMiddleware := middleware.Auth(logger, jwtSecret)
	mux.Handle("POST /api/v1/teams", authMiddleware(http.HandlerFunc(teamHandler.CreateTeam)))
	mux.Handle("GET /api/v1/teams", authMiddleware(http.HandlerFunc(teamHandler.GetUserTeams)))
	mux.Handle("POST /api/v1/teams/{id}/invite", authMiddleware(http.HandlerFunc(teamHandler.InviteUser)))

	mux.Handle("POST /api/v1/tasks", authMiddleware(http.HandlerFunc(taskHandler.CreateTask)))
	mux.Handle("GET /api/v1/tasks", authMiddleware(http.HandlerFunc(taskHandler.GetTasks)))
	mux.Handle("GET /api/v1/tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.GetTaskByID)))
	mux.Handle("PUT /api/v1/tasks/{id}", authMiddleware(http.HandlerFunc(taskHandler.UpdateTask)))
	mux.Handle("GET /api/v1/tasks/{id}/history", authMiddleware(http.HandlerFunc(taskHandler.GetTaskHistory)))

	mux.Handle("POST /api/v1/tasks/{id}/comments", authMiddleware(http.HandlerFunc(commentHandler.CreateComment)))
	mux.Handle("GET /api/v1/tasks/{id}/comments", authMiddleware(http.HandlerFunc(commentHandler.GetComments)))

	return middleware.Logging(logger)(mux)
}
