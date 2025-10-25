package worker

import (
	"cube/task"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Db        map[uuid.UUID]*task.Task
	TaskCount uint64
	Queue     queue.Queue
	Name      string
}

func (w *Worker) RunTask() task.DockerResult {
	t := w.Queue.Dequeue()

	if t == nil {
		log.Println("No tasks in the queue")
		return task.DockerResult{Error: nil}
	}

	taskQueued := t.(task.Task)

	taskPersisted := w.Db[taskQueued.ID]
	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = &taskQueued
	}

	var result task.DockerResult

	if task.ValidStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskPersisted.State {
		case task.Scheduled:
			result = w.StartTask(*taskPersisted)
		case task.Completed:
			result = w.StopTask(*taskPersisted)
		default:
			result.Error = errors.New("we should not be here tf")
		}
	} else {
		err := fmt.Errorf("Invalid transitioning task state from %v to %v", taskPersisted.State, taskQueued.State)
		result.Error = err
	}

	return result
}

func (w *Worker) StartTask(t task.Task) task.DockerResult {
	t.StartTime = time.Now().UTC()
	config := task.NewConfig(&t)
	d := task.NewDocker(config)

	res := d.Run()

	if res.Error != nil {
		log.Printf("Error starting container with ID: %s, %v\n", t.ContainerId, res.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return res
	}

	t.State = task.Running
	t.ContainerId = res.ContainerId
	w.Db[t.ID] = &t

	return res
}

func (w *Worker) StopTask(t task.Task) task.DockerResult {
	config := task.NewConfig(&t)
	d := task.NewDocker(config)

	res := d.Stop(t.ContainerId)

	if res.Error != nil {
		log.Printf("Error stopping container with ID: %s, %v\n", t.ContainerId, res.Error)
		return res
	}

	t.EndTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed container with id %v and Task with id %v\n", t.ContainerId, t.ID)

	return res
}

func (w *Worker) GetTasks() []task.Task {
	res := []task.Task{}

	for _, t := range w.Db {
		res = append(res, *t)
	}

	return res
}

func (w *Worker) CollectStats() {
	fmt.Println("Collecting the stats for this worker")
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}
