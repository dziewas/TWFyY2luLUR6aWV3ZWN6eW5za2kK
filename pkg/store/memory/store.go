package memory

import (
	"context"
	"sync"

	"crawler/pkg/model"
	"crawler/pkg/util"
)

type task struct {
	Id       int
	Url      string
	Interval int
	Attempts []*attempt
}

type attempt struct {
	Response  string
	CreatedAt int64
	Duration  float64
}

type Memory struct {
	tasks map[int]*task
	mutex sync.Mutex
}

func NewMemory() *Memory {
	return &Memory{
		tasks: make(map[int]*task),
	}
}

func (m *Memory) Create(ctx context.Context, t *model.Task) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	id := util.NewID()

	m.tasks[id] = &task{
		Id:       id,
		Url:      t.Url,
		Interval: t.Interval,
	}

	return id, nil
}

func (m *Memory) Get(ctx context.Context, id int) (*model.Task, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	t, found := m.tasks[id]
	if !found {
		return nil, util.ErrResourceNotFound
	}

	return &model.Task{
		Id:       t.Id,
		Url:      t.Url,
		Interval: t.Interval,
	}, nil
}

func (m *Memory) Delete(ctx context.Context, id int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.tasks, id)

	return nil
}

func (m *Memory) ListTasks(ctx context.Context) ([]*model.Task, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tasks := make([]*model.Task, 0, len(m.tasks))

	for _, v := range m.tasks {
		tasks = append(tasks, &model.Task{
			Id:       v.Id,
			Url:      v.Url,
			Interval: v.Interval,
		})
	}

	return tasks, nil
}

func (m *Memory) AddAttempt(ctx context.Context, id int, a *model.Attempt) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	t, found := m.tasks[id]
	if !found {
		return util.ErrResourceNotFound
	}

	t.Attempts = append(t.Attempts, &attempt{
		Response:  a.Response,
		CreatedAt: a.CreatedAt,
		Duration:  a.Duration,
	})

	return nil
}

func (m *Memory) GetAttempts(ctx context.Context, id int) ([]*model.Attempt, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	t, found := m.tasks[id]
	if !found {
		return nil, util.ErrResourceNotFound
	}

	attempts := make([]*model.Attempt, 0, len(t.Attempts))
	for _, a := range t.Attempts {
		attempts = append(attempts, &model.Attempt{
			Response:  a.Response,
			CreatedAt: a.CreatedAt,
			Duration:  a.Duration,
		})
	}

	return attempts, nil
}
