package manager

import (
	"cube/task"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (a *Api) StartTaskHandler(w http.ResponseWriter, r *http.Request) {

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	te := task.TaskEvent{}
	err := d.Decode(&te)
	if err != nil {
		msg := fmt.Sprintf("[Manager] Error decoding task event %v\n", err)
		log.Printf("%s", msg)
		w.WriteHeader(400)
		errMsg := ErrResponse{
			HTTPStatusCode: 400,
			Message:        msg,
		}

		json.NewEncoder(w).Encode(errMsg)

		return
	}

	a.Manager.AddTask(te)
	log.Printf("[Manager] Added task: %v\n", te.Task.ID)
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(te.Task)
}

func (a *Api) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(a.Manager.GetTasks())
}

func (a *Api) StopTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	if taskID == "" {
		log.Printf("[Manager] no task ID passed into the request\n")
		w.WriteHeader(400)
		return
	}

	tID, _ := uuid.Parse(taskID)
	taskToStop, ok := a.Manager.TasksDb[tID]

	if !ok {
		msg := fmt.Sprintf("[Manager] No task found with id %v to stop", tID)
		log.Printf("%s", msg)
		w.WriteHeader(404)
		eResponse := ErrResponse{
			HTTPStatusCode: 404,
			Message:        msg,
		}

		json.NewEncoder(w).Encode(eResponse)
		return
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Completed,
		Timestamp: time.Now(),
	}

	taskCopy := *taskToStop
	taskCopy.State = task.Completed
	te.Task = taskCopy

	a.Manager.AddTask(te)

	log.Printf("[Manager] added task event %v to stop task %v\n", te.ID, tID)
	w.WriteHeader(204)
}
