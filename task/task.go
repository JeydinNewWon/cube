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
