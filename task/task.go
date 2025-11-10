package task

import (
	"context"
	"fmt"
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
	ContainerId   string
	Name          string
	State         TaskState
	Image         string
	CPU           float64
	Memory        int64
	Disk          int64
	ExposedPorts  nat.PortSet
	HostPorts     nat.PortMap
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	EndTime       time.Time
	HealthCheck   string
	RestartCount  int
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
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	CPU           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

type Docker struct {
	Client *client.Client
	Config Config
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

type DockerInspectResponse struct {
	Error     error
	Container *container.InspectResponse
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()

	imageExists, err := checkImageExists(d.Client, d.Config.Image)
	if err != nil {
		log.Printf("Error checking image %s from the image list %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	if !imageExists {
		reader, err := d.Client.ImagePull(ctx, d.Config.Image, image.PullOptions{})
		if err != nil {
			log.Printf("Error pulling the image %s: %v\n", d.Config.Image, err)
			return DockerResult{Error: err}
		}

		io.Copy(os.Stdout, reader)
	}

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
		ExposedPorts: d.Config.ExposedPorts,
		Tty:          false,
	}

	hc := container.HostConfig{
		PublishAllPorts: true,
		RestartPolicy:   rp,
		Resources:       r,
	}

	containerExists, res, err := checkContainerExists(d.Client, d.Config.Name)
	if err != nil {
		log.Printf("Error checking if container existed with name %s\n", d.Config.Name)
		return DockerResult{Error: err}
	}

	var cID string
	if containerExists {
		err := d.Client.ContainerRestart(ctx, res.ID, container.StopOptions{})
		if err != nil {
			log.Printf("Error restarting container with ID %v and name %s\n", res.ID, d.Config.Name)
			return DockerResult{Error: err}
		}
		cID = res.ID
		log.Printf("Restarted container with ID %v\n", cID)

		return DockerResult{ContainerId: cID, Action: "restart", Result: "success"}

	} else {
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

		cID = res.ID
	}

	out, err := d.Client.ContainerLogs(ctx, cID, container.LogsOptions{ShowStderr: true, ShowStdout: true})

	if err != nil {
		log.Printf("Error getting logs for the container %s: %v\n", res.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{ContainerId: cID, Action: "start", Result: "success"}
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

func (d *Docker) Inspect(containerID string) DockerInspectResponse {
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	ctx := context.Background()
	res, err := dc.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Printf("Error inspecting container: %v\n", err)
		return DockerInspectResponse{Error: err}
	}

	return DockerInspectResponse{
		Container: &res,
	}
}

func NewConfig(task *Task) Config {
	return Config{
		Name:          task.Name,
		Image:         task.Image,
		RestartPolicy: task.RestartPolicy,
		CPU:           task.CPU,
		Memory:        task.Memory,
		Disk:          task.Disk,
		ExposedPorts:  task.ExposedPorts,
	}
}

func NewDocker(config Config) Docker {
	newC, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		fmt.Printf("Error creating a new Docker struct: %v\n", err)
		return Docker{
			Client: nil,
			Config: Config{},
		}
	}

	return Docker{
		Client: newC,
		Config: config,
	}
}

func checkImageExists(cli *client.Client, imageName string) (bool, error) {
	ctx := context.Background()

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName || tag == imageName+":latest" {
				return true, nil
			}
		}
	}

	return false, nil
}

func checkContainerExists(cli *client.Client, containerName string) (bool, container.Summary, error) {
	ctx := context.Background()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})

	if err != nil {
		return false, container.Summary{}, err
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if name == "/"+containerName || name == containerName {
				return true, c, nil
			}
		}
	}

	return false, container.Summary{}, err
}
