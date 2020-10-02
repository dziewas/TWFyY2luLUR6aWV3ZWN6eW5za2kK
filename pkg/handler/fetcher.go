package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"crawler/pkg/model"
	"crawler/pkg/store"
	"crawler/pkg/util"
)

const (
	defaultTickerInterval = time.Second * 1
	defaultTimeout        = time.Second * 5
	defaultWorkers        = 10
)

type assignment struct {
	task   *model.Task
	result *model.Attempt
}

type Fetcher struct {
	storage store.Store
}

func NewFetcher(storage store.Store) *Fetcher {
	return &Fetcher{storage: storage}
}

func (f *Fetcher) getTasks(ctx context.Context) []*model.Task {
	tasks, err := f.storage.ListTasks(ctx)
	if err != nil {
		log.Printf("retrieving current tasks from DB failed: %s", err)
		return nil
	}

	now := time.Now().Unix()

	var dueTasks []*model.Task
	for i, task := range tasks {
		attempts, err := f.storage.ListAttempts(ctx, task.Id)
		if err != nil {
			log.Printf("retrieving history for task %d from DB failed: %s", task.Id, err)
			continue
		}

		if len(attempts) == 0 {
			dueTasks = append(dueTasks, tasks[i])
			continue
		}

		lastAttempt := attempts[len(attempts)-1]

		if now-lastAttempt.CreatedAt > int64(task.Interval) {
			dueTasks = append(dueTasks, tasks[i])
		}
	}

	return dueTasks
}

func (f *Fetcher) retriever(finish chan bool, ticker *time.Ticker, assignments chan *assignment) func() {
	ctx := context.Background()

	for {
		select {
		case <-ticker.C:
			tasks := f.getTasks(ctx)
			log.Printf("tasks in this interval: %d\n", len(tasks))
			for i, _ := range tasks {
				assignments <- &assignment{task: tasks[i], result: nil}
			}
		case <-finish:
			break
		}
	}
}

func (f *Fetcher) saver(finish chan bool, results chan *assignment) func() {
	ctx := context.Background()

	for {
		select {
		case result := <-results:
			err := f.storage.AddAttempt(ctx, result.task.Id, result.result)
			if err != nil {
				log.Printf("saving attempt for task %d failed: %s", result.task.Id, err)
			}
		case <-finish:
			break
		}
	}
}

func fetchUrl(url string) string {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("creating request for url '%s' failed: %s", url, err)
		return ""
	}

	httpClient := http.DefaultClient

	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("fetching url '%s' failed: %s", url, err.Error())
		return ""
	}
	defer util.MustClose(res.Body)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("reading body failed from url '%s' failed: %s", url, err)
		return ""
	}

	return string(body)
}

func (f *Fetcher) worker(finish chan bool, assignmentsIn chan *assignment, assignmentsOut chan *assignment) func() {
	for {
		select {
		case a := <-assignmentsIn:
			start := time.Now()
			response := fetchUrl(a.task.Url)
			end := time.Now()

			a.result = &model.Attempt{
				Response:  response,
				CreatedAt: end.Unix(),
				Duration:  end.Sub(start).Seconds(),
			}

			assignmentsOut <- a

		case <-finish:
			break
		}
	}
}

func (f *Fetcher) Start() func() {
	finish := make(chan bool)
	ticker := time.NewTicker(defaultTickerInterval)

	tasks := make(chan *assignment)
	results := make(chan *assignment)
	for i := 0; i < defaultWorkers; i++ {
		go f.worker(finish, tasks, results)
	}
	go f.saver(finish, results)
	go f.retriever(finish, ticker, tasks)

	return func() {
		ticker.Stop()
		finish <- true
	}
}

func (f *Fetcher) Create(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	var task model.Task

	err := d.Decode(&task)
	if err != nil {
		util.EmitHttpError(w, util.ErrValidation)
		return
	}

	defer util.MustClose(r.Body)

	id, err := f.storage.Create(r.Context(), &task)
	if err != nil {
		util.EmitHttpError(w, err)
		return
	}

	w.Header().Add("Location", strconv.Itoa(id))
}

func (f *Fetcher) Delete(w http.ResponseWriter, r *http.Request) {
	ids, ok := mux.Vars(r)["id"]
	if !ok {
		util.EmitHttpError(w, util.ErrValidation)
		return
	}

	id, err := strconv.Atoi(ids)
	if err != nil {
		util.EmitHttpError(w, util.ErrValidation)
		return
	}

	task, err := f.storage.Get(r.Context(), id)
	if err != nil {
		util.EmitHttpError(w, err)
		return
	}

	err = f.storage.Delete(r.Context(), task.Id)
	if err != nil {
		util.EmitHttpError(w, err)
		return
	}
}

func (f *Fetcher) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := f.storage.ListTasks(r.Context())
	if err != nil {
		util.EmitHttpError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(&tasks)
	if err != nil {
		util.EmitHttpError(w, err)
		return
	}
}

func (f *Fetcher) History(w http.ResponseWriter, r *http.Request) {
	ids, ok := mux.Vars(r)["id"]
	if !ok {
		util.EmitHttpError(w, util.ErrValidation)
		return
	}

	id, err := strconv.Atoi(ids)
	if err != nil {
		util.EmitHttpError(w, util.ErrValidation)
		return
	}

	attempts, err := f.storage.ListAttempts(r.Context(), id)
	if err != nil {
		util.EmitHttpError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(&attempts)
	if err != nil {
		util.EmitHttpError(w, err)
		return
	}
}
