package usecase

import (
	"context"

	"github.com/fangimal/TeamTask/internal/domain"
)

type AnalyticsUseCase struct {
	analytics domain.AnalyticsRepository
}

func NewAnalyticsUseCase(analytics domain.AnalyticsRepository) *AnalyticsUseCase {
	return &AnalyticsUseCase{
		analytics: analytics,
	}
}

func (useCase *AnalyticsUseCase) GetTeamStats(ctx context.Context) ([]*domain.TeamStats, error) {
	return useCase.analytics.GetTeamStats(ctx)
}

func (useCase *AnalyticsUseCase) GetTopUsersPerTeam(ctx context.Context) ([]*domain.TopUser, error) {
	return useCase.analytics.GetTopUsersPerTeam(ctx)
}

func (useCase *AnalyticsUseCase) GetIntegrityViolations(ctx context.Context) ([]*domain.IntegrityViolation, error) {
	return useCase.analytics.GetIntegrityViolations(ctx)
}
