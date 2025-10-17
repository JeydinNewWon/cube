package task

import (
	"context"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type TaskState int

const (
	Pending = iota
	Scheduled
	Running
	Completed
	Failed
)

type Task struct {
	ID            uuid.UUID
	ContainerID   uuid.UUID
	Name          string
	State         TaskState
	Image         string
	CPU           float64
	Memory        int64
	Disk          int64
	ExposedPorts  nat.PortSet
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	EndTime       time.Time
}

type TaskEvent struct {
	ID        uuid.UUID
	State     TaskState
	Timestamp time.Time
	Task      Task
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPort   nat.PortSet
	Cmd           []string
	Image         string
	CPU           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

type Docker struct {
	Client client.Client
	Config Config
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, image.PullOptions{})
	if err != nil {
		log.Printf("Error pulling the image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	io.Copy(os.Stdout, reader)

	rp := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}

	r := container.Resources{
		NanoCPUs: int64(d.Config.CPU * math.Pow(10, 9)),
		Memory:   d.Config.Memory,
	}

	cc := container.Config{
		Env:          d.Config.Env,
		Image:        d.Config.Image,
		ExposedPorts: d.Config.ExposedPort,
		Tty:          false,
	}

	hc := container.HostConfig{
		PublishAllPorts: true,
		RestartPolicy:   rp,
		Resources:       r,
	}

	res, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)

	if err != nil {
		log.Printf("Failed to create container with Image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(ctx, res.ID, container.StartOptions{})

	if err != nil {
		log.Printf("Failed to start container with Image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	out, err := d.Client.ContainerLogs(ctx, res.ID, container.LogsOptions{ShowStderr: true, ShowStdout: true})

	if err != nil {
		log.Printf("Error getting logs for the container %s: %v\n", res.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{ContainerId: res.ID, Action: "start", Result: "success"}
}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Attempting to stop container: %s\n", id)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})

	if err != nil {
		log.Printf("Failed to stop the container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, id, container.RemoveOptions{})

	if err != nil {
		log.Printf("Failed to remove the container %s: %v\n", id, err)
		return DockerResult{Error: err}
	}

	return DockerResult{Error: nil, Action: "stop", ContainerId: id, Result: "success"}

}
