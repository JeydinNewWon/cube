package main

import (
	"cube/task"
	"cube/worker"
	"fmt"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	db := make(map[uuid.UUID]*task.Task)
	w := worker.Worker{
		Db:    db,
		Queue: *queue.New(),
	}

	t := task.Task{
		ID:    uuid.New(),
		Name:  "test-container-1",
		State: task.Scheduled,
		Image: "strm/helloworld-http",
	}

	fmt.Println("Starting new task")

	w.AddTask(t)

	result := w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}

	t.ContainerId = result.ContainerId
	fmt.Printf("task %s is running in container %s\n", t.ID, t.ContainerId)

	fmt.Println("Sleepy time")
	time.Sleep(time.Second * 30)

	fmt.Println("Stopping task")
	t.State = task.Completed
	w.AddTask(t)

	result = w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}

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
