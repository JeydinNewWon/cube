package worker

import (
	"cube/task"
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Db        map[uuid.UUID]*task.Task
	TaskCount uint64
	Queue     queue.Queue
	Name      string
}

func (w *Worker) RunTask() {
	fmt.Println("This is the worker")
}

func (w *Worker) StartTask() {
	fmt.Println("This is for starting the task on this worker")
}

func (w *Worker) StopTask() {
	fmt.Println("This is for stopping the task on this worker")
}

func (w *Worker) CollectStats() {
	fmt.Println("Collecting the stats for this worker")
}
