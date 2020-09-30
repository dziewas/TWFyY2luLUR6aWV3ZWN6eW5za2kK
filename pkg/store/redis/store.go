package redis

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"

	"crawler/pkg/model"
	"crawler/pkg/util"
)

const (
	taskPrefix     = "task:"
	responsePrefix = "response:"
	urlKey         = "url"
	intervalKey    = "interval"
	bodyKey        = "body"
	durationKey    = "duration"
	createdAtKey   = "createdAt"

	removeAll = 0
	lastElem  = -1

	historyLimit = 100
)

var (
	taskKeys     = []string{urlKey, intervalKey, createdAtKey}
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

	_, err := s.client.HSet(ctx, task, urlKey, t.Url, intervalKey, t.Interval).Result()
	if err != nil {
		return id, util.Wrap(err, "task create failed")
	}

	_, err = s.client.LPush(ctx, tasks, task).Result()
	if err != nil {
		return id, util.Wrap(err, "adding task to the list failed")
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

	n, err := s.client.LRem(ctx, tasks, removeAll, task).Result()
	if err != nil {
		return util.Wrap(err, "task removal from list failed")
	}

	if n != 1 {
		log.Printf("unexpected number of tasks removed from list %d\n", n)
		return nil
	}

	_, err = s.client.HDel(ctx, task, taskKeys...).Result()
	if err != nil {
		return util.Wrap(err, "task removal failed")
	}

	return nil
}

func (s *Store) ListTasks(ctx context.Context) ([]*model.Task, error) {
	tasks, err := s.client.LRange(ctx, taskPrefix, 0, lastElem).Result()
	if err != nil {
		return nil, util.Wrap(err, "getting list of tasks failed")
	}

	results := make([]*model.Task, 0, len(tasks))

	for _, task := range tasks {
		id, err := strconv.Atoi(strings.Split(task, ":")[1])
		if err != nil {
			return nil, util.Wrap(err, "task with invalid id on the list")
		}

		properties, err := s.client.HGetAll(ctx, task).Result()
		if err != nil {
			return nil, util.Wrap(err, "fetching task failed")
		}

		if len(properties) == 0 {
			log.Printf("unexpected empty task %s", task)
			return nil, nil
		}

		interval, err := strconv.Atoi(properties[intervalKey])
		if err != nil {
			return nil, util.Wrap(err, "interval conversion failed")
		}

		results = append(results, &model.Task{
			Id:       id,
			Url:      properties[urlKey],
			Interval: interval,
		})
	}

	return results, nil
}

func (s *Store) taskExists(ctx context.Context, id int) bool {
	task := taskPrefix + strconv.Itoa(id)
	found, err := s.client.HExists(ctx, task, urlKey).Result()
	if err != nil {
		return false
	}

	return found
}

func (s *Store) AddAttempt(ctx context.Context, id int, a *model.Attempt) error {
	if !s.taskExists(ctx, id) {
		return util.ErrResourceNotFound
	}

	response := responsePrefix + strconv.FormatInt(a.CreatedAt, 10)
	_, err := s.client.HSet(ctx, response, bodyKey, a.Response, durationKey, a.Duration, createdAtKey, a.CreatedAt).Result()
	if err != nil {
		return util.Wrap(err, "response create failed")
	}

	taskResponses := responsePrefix + strconv.Itoa(id)
	_, err = s.client.LPush(ctx, taskResponses, response).Result()
	if err != nil {
		return util.Wrap(err, "adding response to the list failed")
	}

	oldResponses, err := s.client.LRange(ctx, taskResponses, historyLimit, lastElem).Result()
	if err != nil {
		return util.Wrap(err, "getting list of task old responses failed")
	}

	for _, resp := range oldResponses {
		_, err := s.client.HDel(ctx, resp, responseKeys...).Result()
		if err != nil {
			return util.Wrap(err, "old response removal failed")
		}
	}

	_, err = s.client.LTrim(ctx, taskResponses, 0, historyLimit-1).Result()
	if err != nil {
		return util.Wrap(err, "response removal from list failed")
	}

	return nil
}

func (s *Store) GetAttempts(ctx context.Context, id int) ([]*model.Attempt, error) {
	if !s.taskExists(ctx, id) {
		return nil, util.ErrResourceNotFound
	}

	taskResponses := responsePrefix + strconv.Itoa(id)
	responses, err := s.client.LRange(ctx, taskResponses, 0, lastElem).Result()
	if err != nil {
		return nil, util.Wrap(err, "getting list of task responses failed")
	}

	results := make([]*model.Attempt, 0, len(responses))

	for _, response := range responses {
		properties, err := s.client.HGetAll(ctx, response).Result()
		if err != nil {
			return nil, util.Wrap(err, "fetching response failed")
		}

		if len(properties) == 0 {
			log.Printf("unexpected empty response for %s from task %s", response, taskResponses)
			continue
		}

		createdAt, err := strconv.ParseInt(properties[createdAtKey], 10, 64)
		if err != nil {
			return nil, util.Wrap(err, "timestamp conversion failed")
		}

		duration, err := strconv.ParseFloat(properties[durationKey], 64)
		if err != nil {
			return nil, util.Wrap(err, "duration conversion failed")
		}

		results = append(results, &model.Attempt{
			Response:  properties[bodyKey],
			CreatedAt: createdAt,
			Duration:  duration,
		})
	}

	return results, nil
}
