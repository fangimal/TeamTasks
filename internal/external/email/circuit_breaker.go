package email

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/sony/gobreaker"
)

type CircuitBreakerSender struct {
	cb     *gobreaker.CircuitBreaker
	sender domain.EmailSender
	logger *slog.Logger
}

type CircuitBreakerSettings struct {
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
}

func DefaultCircuitBreakerSettings() CircuitBreakerSettings {
	return CircuitBreakerSettings{
		MaxRequests: 2,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
	}
}

func NewCircuitBreakerSender(sender domain.EmailSender, logger *slog.Logger, settings CircuitBreakerSettings) *CircuitBreakerSender {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "email-service",
		MaxRequests: settings.MaxRequests,
		Interval:    settings.Interval,
		Timeout:     settings.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Warn("circuit breaker state changed",
				slog.String("name", name),
				slog.String("from", from.String()),
				slog.String("to", to.String()),
			)
		},
	})

	return &CircuitBreakerSender{
		cb:     cb,
		sender: sender,
		logger: logger,
	}
}

func (sender *CircuitBreakerSender) SendInviteEmail(ctx context.Context, email string, teamName string) error {
	_, err := sender.cb.Execute(func() (interface{}, error) {
		return nil, sender.sender.SendInviteEmail(ctx, email, teamName)
	})
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			sender.logger.WarnContext(ctx, "circuit breaker open, email not sent",
				slog.String("email", email),
				slog.String("team_name", teamName),
			)
			return domain.ErrCircuitBreakerOpen
		}
		return err
	}
	return nil
}
