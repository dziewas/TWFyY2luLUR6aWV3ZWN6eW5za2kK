// +build unit !integration

package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"crawler/pkg/model"
	"crawler/pkg/util"
)

func create(t *testing.T, ctx context.Context, store *Memory, task *model.Task) int {
	id, err := store.Create(ctx, task)
	require.NoError(t, err)

	return id
}

func TestCreate(t *testing.T) {
	store := NewMemory()

	ctx := context.Background()

	id1 := create(t, ctx, store, &model.Task{})
	id2 := create(t, ctx, store, &model.Task{})

	assert.NotEqual(t, id1, id2)
}

func TestGetExisting(t *testing.T) {
	store := NewMemory()

	newTask := &model.Task{
		Url:      "http://example.com",
		Interval: 60,
	}

	ctx := context.Background()

	id := create(t, ctx, store, newTask)
	newTask.Id = id

	readTask, err := store.Get(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, newTask, readTask)

}

func TestGetNonExisting(t *testing.T) {
	store := NewMemory()

	ctx := context.Background()

	_, err := store.Get(ctx, 123)
	assert.True(t, errors.Is(err, util.ErrResourceNotFound))
}

func TestDeleteExisting(t *testing.T) {
	store := NewMemory()

	ctx := context.Background()

	id := create(t, ctx, store, &model.Task{})
	err := store.Delete(ctx, id)
	require.NoError(t, err)

	_, err = store.Get(ctx, id)
	assert.True(t, errors.Is(err, util.ErrResourceNotFound))
}

func TestDeleteNonExisting(t *testing.T) {
	store := NewMemory()

	ctx := context.Background()

	err := store.Delete(ctx, 123)
	require.NoError(t, err)
}

func TestListTasks(t *testing.T) {
	store := NewMemory()

	ctx := context.Background()

	newTasks := []*model.Task{
		{
			Url:      "http://example.com",
			Interval: 60,
		},
		{
			Url:      "http://dummy.com",
			Interval: 10,
		}}

	newTasks[0].Id = create(t, ctx, store, newTasks[0])
	newTasks[1].Id = create(t, ctx, store, newTasks[1])

	readTasks, err := store.ListTasks(ctx)
	require.NoError(t, err)
	assert.Equal(t, newTasks, readTasks)
}

func TestAddAttempts(t *testing.T) {
	store := NewMemory()

	ctx := context.Background()

	id := create(t, ctx, store, &model.Task{})

	err := store.AddAttempt(ctx, id, &model.Attempt{})
	require.NoError(t, err)
}

func TestGetAttempts(t *testing.T) {
	store := NewMemory()

	ctx := context.Background()

	task1 := &model.Task{
		Url:      "http://example.com",
		Interval: 60,
	}

	task2 := &model.Task{
		Url:      "http://dummy.com",
		Interval: 10,
	}

	task1.Id = create(t, ctx, store, task1)
	task2.Id = create(t, ctx, store, task2)

	attempt1 := &model.Attempt{
		Response:  "response1",
		CreatedAt: 123,
		Duration:  1,
	}

	attempt2 := &model.Attempt{
		Response:  "response2",
		CreatedAt: 543,
		Duration:  2,
	}

	err := store.AddAttempt(ctx, task1.Id, attempt1)
	require.NoError(t, err)

	err = store.AddAttempt(ctx, task1.Id, attempt2)
	require.NoError(t, err)

	err = store.AddAttempt(ctx, task2.Id, attempt2)
	require.NoError(t, err)

	task1Attempts, err := store.ListAttempts(ctx, task1.Id)
	require.NoError(t, err)
	assert.Equal(t, []*model.Attempt{attempt1, attempt2}, task1Attempts)

	task2Attempts, err := store.ListAttempts(ctx, task2.Id)
	require.NoError(t, err)
	assert.Equal(t, []*model.Attempt{attempt2}, task2Attempts)
}
