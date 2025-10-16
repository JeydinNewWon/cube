package worker

import (
	"cube/task"
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	db        map[uuid.UUID]*task.Task
	taskCount uint64
	queue     queue.Queue
	name      string
}

func runTask(worker *Worker) {
	fmt.Println("This is the worker %v", worker)
}

func startTask(worker *Worker) {
	fmt.Println("This is for starting the task on this worker")
}

func stopTask(worker *Worker) {
	fmt.Println("This is for stopping the task on this worker")
}

func collectStats(worker *Worker) {
	fmt.Println("Collecting the stats for this worker")
}
