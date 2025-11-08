package main

import (
	"cube/manager"
	"cube/task"
	"cube/worker"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	whost := os.Getenv("CUBE_WORKER_HOST")
	wport, _ := strconv.Atoi(os.Getenv("CUBE_WORKER_PORT"))

	mhost := os.Getenv("CUBE_MANAGER_HOST")
	mport, _ := strconv.Atoi(os.Getenv("CUBE_MANAGER_PORT"))

	fmt.Println("Starting Cube worker")

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
	wapi := worker.Api{Address: whost, Port: wport, Worker: &w}

	go wapi.Start()

	time.Sleep(3 * time.Second)

	go w.RunTasks()
	go w.CollectStats()

	log.Println("Starting Cube Manager")

	workers := []string{fmt.Sprintf("%s:%d", whost, wport)}
	m := manager.New(workers)
	mapi := manager.Api{
		Address: mhost,
		Port:    mport,
		Manager: m,
	}

	go mapi.Start()

	time.Sleep(2 * time.Second)

	go m.ProcessTasks()
	go m.UpdateTasks()

	select {}

	// go func() {
	// 	for {
	// 		fmt.Printf("[Manager] Updating tasks from %d workers\n", len(m.Workers))
	// 		m.UpdateTasks()
	// 		time.Sleep(15 * time.Second)
	// 	}
	// }()

	// for {
	// 	for _, t := range m.TasksDb {
	// 		fmt.Printf("[Manager] Task id: %s, state: %d\n", t.ID, t.State)
	// 		time.Sleep(15 * time.Second)
	// 	}
	// }
	// db := make(map[uuid.UUID]*task.Task)
	// w := worker.Worker{
	// 	Db:    db,
	// 	Queue: *queue.New(),
	// }

	// t := task.Task{
	// 	ID:    uuid.New(),
	// 	Name:  "test-container-1",
	// 	State: task.Scheduled,
	// 	Image: "strm/helloworld-http",
	// }

	// fmt.Println("Starting new task")

	// w.AddTask(t)

	// result := w.RunTask()
	// if result.Error != nil {
	// 	panic(result.Error)
	// }

	// t.ContainerId = result.ContainerId
	// fmt.Printf("task %s is running in container %s\n", t.ID, t.ContainerId)

	// fmt.Println("Sleepy time")
	// time.Sleep(time.Second * 30)

	// fmt.Println("Stopping task")
	// t.State = task.Completed
	// w.AddTask(t)

	// result = w.RunTask()
	// if result.Error != nil {
	// 	panic(result.Error)
	// }

}

// func createContainer() (*task.Docker, *task.DockerResult) {

// 	c := task.Config{
// 		Name:  "test-container-1",
// 		Image: "postgres:13",
// 		Env: []string{
// 			"POSTGRES_USER=cube",
// 			"POSTGRES_PASSWORD=secret",
// 		},
// 	}

// 	dc, _ := client.NewClientWithOpts(client.FromEnv)
// 	d := task.Docker{
// 		Client: dc,
// 		Config: c,
// 	}

// 	res := d.Run()

// 	if res.Error != nil {
// 		fmt.Printf("Error: %v\n", res.Error)
// 		return nil, nil
// 	}

// 	fmt.Printf("Container %s is running with config %v\n", res.ContainerId, c)

// 	return &d, &res

// }

// func stopContainer(d *task.Docker, id string) *task.DockerResult {
// 	result := d.Stop(id)
// 	if result.Error != nil {
// 		fmt.Printf("%v\n", result.Error)
// 		return nil
// 	}

// 	fmt.Printf(
// 		"Container %s has been stopped and removed\n", result.ContainerId)
// 	return &result
// }
