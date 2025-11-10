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
	Stats     *Stats
}

func (w *Worker) runTask() task.DockerResult {
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
		switch taskQueued.State {
		case task.Scheduled:
			result = w.StartTask(taskQueued)
		case task.Completed:
			result = w.StopTask(taskQueued)
		default:
			result.Error = errors.New("we should not be here tf")
		}
	} else {
		err := fmt.Errorf("invalid transitioning task state from %v to %v", taskPersisted.State, taskQueued.State)
		result.Error = err
	}

	return result
}

func (w *Worker) updateTasks() {
	for id, t := range w.Db {
		if t.State == task.Running {
			res := w.InspectTask(*t)
			if res.Error != nil {
				log.Printf("Error with updating task through inspection %v\n", res.Error)
			}

			if res.Container == nil {
				log.Printf("No container found for running task %v\n", id)
				w.Db[id].State = task.Failed
			}

			if res.Container.State.Status == "exited" {
				log.Printf("Container for task %s is not in running state %s\n", id, res.Container.State.Status)
				w.Db[id].State = task.Failed
			}

			w.Db[id].HostPorts = res.Container.NetworkSettings.NetworkSettingsBase.Ports

		}
	}
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

	log.Printf("[Worker] started task %v\n", t.ID)

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

func (w *Worker) GetTasks() []*task.Task {
	res := []*task.Task{}

	for _, t := range w.Db {
		res = append(res, t)
	}

	return res
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) CollectStats() {
	for {
		log.Println("Collecting stats")
		w.Stats = GetStats()
		w.Stats.TaskCount = w.TaskCount
		time.Sleep(15 * time.Second)
	}
}

func (w *Worker) RunTasks() {
	for {
		if w.Queue.Len() != 0 {
			result := w.runTask()
			if result.Error != nil {
				log.Printf("Error running task: %v\n", result.Error)
			}
		} else {
			log.Printf("No tasks to process currently.\n")
		}
		log.Println("Sleeping for 10 seconds.")
		time.Sleep(10 * time.Second)
	}
}

func (w *Worker) UpdateTasks() {
	for {
		log.Println("Checking status of tasks")
		w.updateTasks()
		log.Println("Task updates completed")
		log.Println("Sleeping for 15 seconds")
		time.Sleep(15 * time.Second)
	}
}

func (w *Worker) InspectTask(t task.Task) task.DockerInspectResponse {
	config := task.NewConfig(&t)
	d := task.NewDocker(config)
	return d.Inspect(t.ContainerId)
}
