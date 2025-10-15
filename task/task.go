package task

import (
	"time"

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
	id            uuid.UUID
	containerID   uuid.UUID
	name          string
	state         TaskState
	image         string
	cpu           float64
	memory        int64
	disk          int64
	exposedPorts  nat.PortSet
	portBindings  map[string]string
	restartPolicy string
	startTime     time.Time
	endTime       time.Time
}

type TaskEvent struct {
	id        uuid.UUID
	state     TaskState
	timestamp time.Time
	task      Task
}
