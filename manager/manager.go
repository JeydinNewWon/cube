package manager

import (
	"cube/task"
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TasksDb       map[string]task.Task
	TaskEventDb   map[string]task.TaskEvent
	Workers       []string
	WorkerTaskMap map[string]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
}

func (manager *Manager) SelectWorker() {
	fmt.Println("I am going to select a worker ts!")
}

func (manager *Manager) UpdateTasks() {
	fmt.Println("Updating the tasks and data for workers")
}

func (manager *Manager) SendWork() {
	fmt.Println("delegating some work to some task...")
}
