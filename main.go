package main

import (
	"cube/manager"
	"cube/node"
	"cube/task"
	"cube/worker"
	"fmt"
	"os"

	"time"

	"github.com/docker/docker/client"
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

	fmt.Printf("Create a test container\n")
	dockerTask, createResult := createContainer()
	if createResult.Error != nil {
		fmt.Printf("%v", createResult.Error)
		os.Exit(1)
	}

	time.Sleep(time.Second * 5)

	fmt.Printf("Stopping container %s\n", createResult.ContainerId)
	_ = stopContainer(dockerTask, createResult.ContainerId)

}

func createContainer() (*task.Docker, *task.DockerResult) {

	c := task.Config{
		Name:  "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=secret",
		},
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv)
	d := task.Docker{
		Client: *dc,
		Config: c,
	}

	res := d.Run()

	if res.Error != nil {
		fmt.Printf("Error: %v\n", res.Error)
		return nil, nil
	}

	fmt.Printf("Container %s is running with config %v\n", res.ContainerId, c)

	return &d, &res

}

func stopContainer(d *task.Docker, id string) *task.DockerResult {
	result := d.Stop(id)
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil
	}

	fmt.Printf(
		"Container %s has been stopped and removed\n", result.ContainerId)
	return &result
}
