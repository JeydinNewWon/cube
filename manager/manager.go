package manager

import (
	"bytes"
	"cube/task"
	"cube/worker"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
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
	if m.LastWorker+1 < len(m.Workers) {
		newWorker = m.LastWorker + 1
		m.LastWorker++
	} else {
		newWorker = 0
		m.LastWorker = 0
	}

	return m.Workers[newWorker]
}

func (m *Manager) updateTasks() {

	for _, w := range m.Workers {
		log.Printf("Checking worker %v for task updates\n", w)
		url := fmt.Sprintf("http://%s/tasks", w)

		res, err := http.Get(url)

		if err != nil {
			log.Printf("[Manager] Error getting tasks info %v\n", err)
			continue
		}

		d := json.NewDecoder(res.Body)
		var tasks []*task.Task
		err = d.Decode(&tasks)

		if err != nil {
			log.Printf("[Manager] error in unmarshalling tasks %v", err)
			continue
		}

		for _, t := range tasks {
			log.Printf("[Manager] Attempting to update task %v\n", t.ID)

			_, ok := m.TasksDb[t.ID]

			if !ok {
				log.Printf("[Manager] Task with ID %v was not found!", t.ID)
				continue
			}

			if m.TasksDb[t.ID].State != t.State {
				m.TasksDb[t.ID].State = t.State
			}

			m.TasksDb[t.ID].StartTime = t.StartTime
			m.TasksDb[t.ID].EndTime = t.EndTime
			m.TasksDb[t.ID].ContainerId = t.ContainerId
			m.TasksDb[t.ID].HostPorts = t.HostPorts

		}

	}

}

func (m *Manager) SendWork() {
	if m.Pending.Len() < 1 {
		log.Printf("[Manager] No pending tasks to allocate")
		return
	}

	newWorker := m.SelectWorker()

	e := m.Pending.Dequeue()
	event := e.(task.TaskEvent)

	t := event.Task
	log.Printf("[Manager] Pulled %v off the pending queue\n", t)

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
		log.Printf("[Manager] Error connecting to %v\n", err)
		m.Pending.Enqueue(event)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Manager] Error sending request %v\n", err)
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

func (m *Manager) GetTasks() []*task.Task {
	tasks := []*task.Task{}
	for _, t := range m.TasksDb {
		tasks = append(tasks, t)
	}

	return tasks
}

func (m *Manager) checkTaskHealth(t task.Task) error {
	log.Printf("Calling health check for task %s: %s\n", t.ID, t.HealthCheck)

	w := m.TaskWorkerMap[t.ID]
	hostPort := getHostPort(t.HostPorts)
	worker := strings.Split(w, ":")
	url := fmt.Sprintf("http://%s:%s%s", worker[0], *hostPort, t.HealthCheck)

	log.Printf("[Manager] Calling health check for task %s: %s\n", t.ID, url)

	resp, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("[Manager] Error connecting to health check %s\n", err)
		log.Println(msg)
		return errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("[Manager] Error health check for task %s did not return 200\n", t.ID)
		log.Println(msg)
		return errors.New(msg)
	}

	log.Printf("[Manager] Task %s health check response: %v\n", t.ID, resp.StatusCode)

	return nil

}

func (m *Manager) doHealthChecks() {
	for _, t := range m.TasksDb {
		if t.RestartCount >= 3 {
			return
		}

		log.Printf("[health check loop] %#v\n", t)

		if t.State == task.Running {
			err := m.checkTaskHealth(*t)
			if err != nil {
				m.restartTask(t)
			}
		} else if t.State == task.Failed {
			m.restartTask(t)
		}
	}
}

func (m *Manager) restartTask(t *task.Task) {
	w := m.TaskWorkerMap[t.ID]
	t.State = task.Scheduled
	t.RestartCount++
	m.TasksDb[t.ID] = t

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Running,
		Timestamp: time.Now(),
		Task:      *t,
	}

	data, err := json.Marshal(te)
	if err != nil {
		log.Printf("[Manager] unable to marshal task object: %v.", t)
		return
	}

	url := fmt.Sprintf("http://%s/tasks", w)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("[Manager] error conntecting to %v: %v", w, err)
		m.Pending.Enqueue(t)
		return
	}

	d := json.NewDecoder(res.Body)

	if res.StatusCode != http.StatusCreated {
		e := worker.ErrResponse{}
		err := d.Decode(&e)
		if err != nil {
			fmt.Printf("[Manager] error decoding response %s\n", err)
			return
		}

		log.Printf("[Manager] Response error (%d): %s", e.HTTPStatusCode, e.Message)
		return
	}

	newTask := task.Task{}
	err = d.Decode(&newTask)
	if err != nil {
		fmt.Printf("[Manager] Error decoding response: %v\n", err)
		return
	}

	log.Printf("%#v\n", t)
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

func (m *Manager) UpdateTasks() {
	for {
		log.Println("[Manager] Checking for any task updates from the workers")
		m.updateTasks()
		log.Println("[Manager] Task updates completed")
		log.Println("[Manager] Sleeping for 15 seconds")
		time.Sleep(15 * time.Second)
	}
}

func (m *Manager) ProcessTasks() {
	for {
		log.Println("[Manager] Processing any tasks in the queue")
		m.SendWork()
		log.Println("[Manager] Sleeping for 10 seconds")
		time.Sleep(10 * time.Second)
	}
}

func (m *Manager) DoHealthChecks() {
	for {
		log.Println("[Manager] Performing task health check")
		m.doHealthChecks()
		log.Println("[Manager] Task health checks completed")
		log.Println("[Manager] Sleeping for 60 seconds")
		time.Sleep(60 * time.Second)
	}
}

func getHostPort(ports nat.PortMap) *string {
	for k := range ports {
		return &ports[k][0].HostPort
	}

	return nil
}
