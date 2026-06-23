package redisrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/fangimal/TeamTask/internal/domain"
)

type taskListWrapper struct {
	Tasks []*domain.Task `json:"tasks"`
	Total int64          `json:"total"`
}

type taskCacheRepository struct {
	client *redis.Client
}

func NewTaskCacheRepository(client *redis.Client) domain.TaskCacheRepository {
	return &taskCacheRepository{client: client}
}

func (repo *taskCacheRepository) Ping(ctx context.Context) error {
	return repo.client.Ping(ctx).Err()
}

func (repo *taskCacheRepository) Get(ctx context.Context, key string) ([]*domain.Task, int64, error) {
	data, err := repo.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, domain.ErrCacheMiss
		}
		return nil, 0, fmt.Errorf("redis get: %w", err)
	}

	var wrapper taskListWrapper
	if err = json.Unmarshal(data, &wrapper); err != nil {
		return nil, 0, fmt.Errorf("redis unmarshal: %w", err)
	}

	return wrapper.Tasks, wrapper.Total, nil
}

func (repo *taskCacheRepository) Set(ctx context.Context, key string, tasks []*domain.Task, total int64, ttl time.Duration) error {
	wrapper := taskListWrapper{Tasks: tasks, Total: total}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("redis marshal: %w", err)
	}

	if err = repo.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (repo *taskCacheRepository) Invalidate(ctx context.Context, pattern string) error {
	var cursor uint64

	for {
		keys, nextCursor, err := repo.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("redis scan: %w", err)
		}

		if len(keys) > 0 {
			if err = repo.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis del: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
