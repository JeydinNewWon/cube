package manager

import (
	"bytes"
	"cube/task"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TasksDb       map[uuid.UUID]*task.Task
	TaskEventDb   map[uuid.UUID]*task.TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int
}

func (m *Manager) SelectWorker() string {
	var newWorker int
	if m.LastWorker < len(m.Workers) {
		newWorker = m.LastWorker + 1
		m.LastWorker++
	} else {
		newWorker = 0
		m.LastWorker = 0
	}

	return m.Workers[newWorker]
}

func (m *Manager) UpdateTasks() {

	for _, w := range m.Workers {
		log.Printf("Checking worker %v for task updates\n", w)
		url := fmt.Sprintf("http://%s/tasks", w)

		res, err := http.Get(url)

		if err != nil {
			log.Printf("Error getting tasks info %v\n", err)
			return
		}

		d := json.NewDecoder(res.Body)
		var tasks []*task.Task
		err = d.Decode(&tasks)

		if err != nil {
			log.Printf("error in unmarshalling tasks %v", err)
			return
		}

		for _, t := range tasks {
			log.Printf("Attempting to update task %v\n", t.ID)

			_, ok := m.TasksDb[t.ID]

			if !ok {
				log.Printf("Task with ID %v was not found!", t.ID)
				return
			}

			if m.TasksDb[t.ID].State != t.State {
				m.TasksDb[t.ID].State = t.State
			}

			m.TasksDb[t.ID].StartTime = t.StartTime
			m.TasksDb[t.ID].EndTime = t.EndTime
			m.TasksDb[t.ID].ContainerId = t.ContainerId

		}

	}

}

func (m *Manager) SendWork() {
	if m.Pending.Len() < 1 {
		log.Printf("No pending tasks to allocate")
		return
	}

	newWorker := m.SelectWorker()

	e := m.Pending.Dequeue()
	event := e.(task.TaskEvent)

	t := event.Task
	log.Printf("Pulled %v off the pending queue\n", t)

	m.TaskEventDb[event.ID] = &event
	m.TaskWorkerMap[t.ID] = newWorker
	m.WorkerTaskMap[newWorker] = append(m.WorkerTaskMap[newWorker], t.ID)

	t.State = task.Scheduled
	m.TasksDb[t.ID] = &t

	data, err := json.Marshal(event)

	if err != nil {
		log.Printf("Unable to marshal task object %v\n", t)
	}

	url := fmt.Sprintf("http://%s/tasks", newWorker)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

	if err != nil {
		log.Printf("Error connecting to %v\n", err)
		m.Pending.Enqueue(event)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error sending request %v\n", err)
		return
	}

	d := json.NewDecoder(resp.Body)

	err = d.Decode(&d)

	if err != nil {
		log.Printf("Error unmarshalling tasks: %s\n", err.Error())
	}

}

func (m *Manager) AddTask(te task.TaskEvent) {
	m.Pending.Enqueue(te)
}

func New(workers []string) *Manager {
	tasksDb := make(map[uuid.UUID]*task.Task)
	taskEventDb := make(map[uuid.UUID]*task.TaskEvent)
	workerTaskMap := make(map[string][]uuid.UUID)
	taskWorkerMap := make(map[uuid.UUID]string)
	for worker := range workers {
		workerTaskMap[workers[worker]] = []uuid.UUID{}
	}

	manager := Manager{
		Pending:       *queue.New(),
		Workers:       workers,
		TasksDb:       tasksDb,
		TaskEventDb:   taskEventDb,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
	}

	return &manager
}
