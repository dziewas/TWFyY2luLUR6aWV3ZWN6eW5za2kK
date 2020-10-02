package store

import (
	"context"

	"crawler/pkg/model"
)

type Store interface {
	Create(ctx context.Context, task *model.Task) (int, error)
	Get(ctx context.Context, id int) (*model.Task, error)
	Delete(ctx context.Context, id int) error
	ListTasks(ctx context.Context) ([]*model.Task, error)
	AddAttempt(ctx context.Context, id int, attempt *model.Attempt) error
	ListAttempts(ctx context.Context, id int) ([]*model.Attempt, error)
}
