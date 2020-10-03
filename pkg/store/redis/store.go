package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"

	"crawler/pkg/model"
	"crawler/pkg/util"
)

const (
	taskPrefix     = "task:"
	responsePrefix = "response:"
	idKey          = "id"
	urlKey         = "url"
	intervalKey    = "interval"
	bodyKey        = "body"
	durationKey    = "duration"
	createdAtKey   = "createdAt"

	removeAll = 0
	lastElem  = -1

	historyLimit = 10
)

var (
	taskKeys     = []string{idKey, urlKey, intervalKey}
	responseKeys = []string{bodyKey, durationKey, createdAtKey}
)

type Store struct {
	client *redis.Client
}

func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) Create(ctx context.Context, t *model.Task) (int, error) {
	id := util.NewID()

	tasks := taskPrefix
	task := taskPrefix + strconv.Itoa(id)

	cmds, err := s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, task, idKey, id, urlKey, t.Url, intervalKey, t.Interval)
		pipe.LPush(ctx, tasks, task)

		return nil
	})

	if err != nil || len(cmds) != 2 {
		return id, util.Wrap(err, "saving task to DB failed")
	}

	return id, nil
}

func (s *Store) Get(ctx context.Context, id int) (*model.Task, error) {
	task := taskPrefix + strconv.Itoa(id)

	properties, err := s.client.HGetAll(ctx, task).Result()
	if err != nil {
		return nil, util.Wrap(err, "task get failed")
	}

	if len(properties) == 0 {
		return nil, util.ErrResourceNotFound
	}

	interval, err := strconv.Atoi(properties[intervalKey])
	if err != nil {
		return nil, util.Wrap(err, "interval conversion failed")
	}

	return &model.Task{
		Id:       id,
		Url:      properties[urlKey],
		Interval: interval,
	}, nil
}

func (s *Store) Delete(ctx context.Context, id int) error {
	tasks := taskPrefix
	task := taskPrefix + strconv.Itoa(id)

	results, err := s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LRem(ctx, tasks, removeAll, task)
		pipe.HDel(ctx, task, taskKeys...)

		return nil
	})

	if err != nil || len(results) != 2 {
		return util.Wrap(err, "deleting task from DB failed")
	}

	return nil
}

func (s *Store) ListTasks(ctx context.Context) ([]*model.Task, error) {
	tasks, err := s.client.LRange(ctx, taskPrefix, 0, lastElem).Result()
	if err != nil {
		return nil, util.Wrap(err, "getting list of tasks failed")
	}

	results, err := s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, task := range tasks {
			pipe.HGetAll(ctx, task)
		}

		return nil
	})

	if err != nil || len(results) != len(tasks) {
		return nil, util.Wrap(err, "getting some tasks from DB failed")
	}

	ret := make([]*model.Task, 0, len(tasks))

	for _, result := range results {
		properties, ok := result.(*redis.StringStringMapCmd)
		if !ok {
			return nil, util.Wrap(err, "fetching task properties failed")
		}

		id, err := strconv.Atoi(properties.Val()[idKey])
		if err != nil {
			return nil, util.Wrap(err, "interval conversion failed")
		}

		interval, err := strconv.Atoi(properties.Val()[intervalKey])
		if err != nil {
			return nil, util.Wrap(err, "interval conversion failed")
		}

		ret = append(ret, &model.Task{
			Id:       id,
			Url:      properties.Val()[urlKey],
			Interval: interval,
		})
	}

	return ret, nil
}

func (s *Store) taskExists(ctx context.Context, id int) bool {
	task := taskPrefix + strconv.Itoa(id)

	return s.client.HExists(ctx, task, urlKey).Val()
}

func (s *Store) AddAttempt(ctx context.Context, id int, a *model.Attempt) error {
	if !s.taskExists(ctx, id) {
		return util.ErrResourceNotFound
	}

	err := s.historyCleanup(ctx, id)
	if err != nil {
		return util.Wrap(err, "history cleanup failed")
	}

	response := fmt.Sprintf("%s%d:%d", responsePrefix, id, a.CreatedAt)
	responses := responsePrefix + strconv.Itoa(id)

	results, err := s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, response, bodyKey, a.Response, durationKey, a.Duration, createdAtKey, a.CreatedAt)
		pipe.RPush(ctx, responses, response)

		return nil
	})

	if err != nil || len(results) != 2 {
		return util.Wrap(err, "saving response to DB failed")
	}

	return nil
}

func (s *Store) historyCleanup(ctx context.Context, id int) error {
	responses := responsePrefix + strconv.Itoa(id)

	historySize := s.client.LLen(ctx, responses).Val()
	if historySize <= historyLimit {
		return nil
	}

	oldResponses, err := s.client.LRange(ctx, responses, 0, historySize-historyLimit-1).Result()
	if err != nil {
		return util.Wrap(err, "getting list of task old responses failed")
	}

	results, err := s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LTrim(ctx, responses, historySize-historyLimit, lastElem)
		for _, resp := range oldResponses {
			pipe.HDel(ctx, resp, responseKeys...)
		}

		return nil
	})

	if err != nil || len(results) != len(oldResponses)+1 {
		return util.Wrap(err, "removing old responses from DB failed")
	}

	return nil
}

func (s *Store) ListAttempts(ctx context.Context, id int) ([]*model.Attempt, error) {
	if !s.taskExists(ctx, id) {
		return nil, util.ErrResourceNotFound
	}

	taskResponses := responsePrefix + strconv.Itoa(id)
	responses, err := s.client.LRange(ctx, taskResponses, 0, lastElem).Result()
	if err != nil {
		return nil, util.Wrap(err, "getting list of task responses failed")
	}

	results, err := s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, response := range responses {
			pipe.HGetAll(ctx, response)
		}

		return nil
	})

	if err != nil || len(results) != len(responses) {
		return nil, util.Wrap(err, "getting some responses from DB failed")
	}

	ret := make([]*model.Attempt, 0, len(responses))

	for _, result := range results {
		properties, ok := result.(*redis.StringStringMapCmd)
		if !ok {
			return nil, util.Wrap(err, "fetching response properties failed")
		}

		createdAt, err := strconv.ParseInt(properties.Val()[createdAtKey], 10, 64)
		if err != nil {
			return nil, util.Wrap(err, "timestamp conversion failed")
		}

		duration, err := strconv.ParseFloat(properties.Val()[durationKey], 64)
		if err != nil {
			return nil, util.Wrap(err, "duration conversion failed")
		}

		ret = append(ret, &model.Attempt{
			Response:  properties.Val()[bodyKey],
			CreatedAt: createdAt,
			Duration:  duration,
		})
	}

	return ret, nil
}
