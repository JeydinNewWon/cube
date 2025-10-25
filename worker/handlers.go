package worker

import (
	"cube/task"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ErrResponse struct {
	HTTPStatusCode int
	Message        string
}

func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	te := task.TaskEvent{}

	err := d.Decode(te)

	if err != nil {
		msg := fmt.Sprintf("Error decoding body: %v\n", err)
		log.Printf(msg)
		w.WriteHeader(400)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}

		json.NewEncoder(w).Encode(e)
		return
	}

	a.Worker.AddTask(te.Task)
	log.Printf("Added task %s\n", te.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(te.Task)
}

func (a *Api) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Worker.GetTasks())

}

func (a *Api) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	if taskID == "" {
		msg := fmt.Sprintf("taskID was not supplied properly or is nil %v", taskID)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}
		log.Println(msg)
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(e)
		return
	}

	taskUUID, err := uuid.Parse(taskID)

	if err != nil {
		msg := fmt.Sprintf("Had issues parsing task ID of %v", taskID)
		e := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}
		w.WriteHeader(400)
		log.Println(msg)
		json.NewEncoder(w).Encode(e)
		return
	}

	_, ok := a.Worker.Db[taskUUID]

	if !ok {
		msg := fmt.Sprintf("Cannot find valid task with ID: %s\n", taskID)
		e := ErrResponse{
			HTTPStatusCode: 404,
			Message:        msg,
		}
		log.Println(msg)
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(e)
		return
	}

	taskToStop := a.Worker.Db[taskUUID]
	taskCopy := *taskToStop
	taskCopy.State = task.Completed
	a.Worker.AddTask(taskCopy)

	log.Printf("Added task %v to stop container %v\n", taskToStop.ID, taskToStop.ContainerId)
	w.WriteHeader(204)
}
