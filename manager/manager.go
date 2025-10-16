package manager

import (
	"cube/task"
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	pending       queue.Queue
	tasksDb       map[string]task.Task
	taskEventDb   map[string]task.TaskEvent
	workers       []string
	workerTaskMap map[string]uuid.UUID
	taskWorkerMap map[uuid.UUID]string
}

func selectWorker(manager *Manager) {
	fmt.Println("I am going to select a worker ts!")
}

func updateTasks(manager *Manager) {
	fmt.Println("Updating the tasks and data for workers")
}

func sendWork(manager *Manager) {
	fmt.Println("delegating some work to some task...")
}
