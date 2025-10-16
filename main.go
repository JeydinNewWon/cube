package main

import (
	"cube/manager"
	"cube/node"
	"cube/task"
	"cube/worker"
	"fmt"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	new_t := task.Task{
		ID:     uuid.New(),
		Name:   "Task-1",
		State:  task.Pending,
		Image:  "Image",
		Memory: 1024,
		Disk:   1,
	}

	task_e := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Pending,
		Timestamp: time.Now(),
		Task:      new_t,
	}

	fmt.Printf("%v\n", new_t)
	fmt.Printf("%v\n", task_e)

	w := worker.Worker{
		Db:        make(map[uuid.UUID]*task.Task),
		TaskCount: 1,
		Queue:     *queue.New(),
		Name:      "Worker-1",
	}

	fmt.Printf("%v\n", w)

	w.StartTask()
	w.RunTask()
	w.StopTask()
	w.CollectStats()

	m := manager.Manager{
		Pending:       *queue.New(),
		TasksDb:       make(map[string]task.Task),
		TaskEventDb:   make(map[string]task.TaskEvent),
		Workers:       []string{w.Name},
		WorkerTaskMap: make(map[string]uuid.UUID),
		TaskWorkerMap: make(map[uuid.UUID]string),
	}

	fmt.Printf("%v\n", m)

	m.UpdateTasks()
	m.SelectWorker()
	m.SendWork()

	n := node.Node{
		Name:   "Node-1",
		Ip:     "192.168.1.1",
		Cores:  4,
		Memory: 1024,
		Disk:   25,
		Role:   "worker",
	}

	fmt.Printf("%v\n", n)
}
