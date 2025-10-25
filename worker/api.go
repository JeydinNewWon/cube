package worker

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Api struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *chi.Mux
}

func (a *Api) InitRouter() {
	a.Router = chi.NewRouter()
	a.Router.Route("/tasks", func(r chi.Router) {
		r.Post("/", a.StartTaskHandler)
		r.Get("/", a.GetTaskHandler)
		r.Route("/{taskID}", func(r chi.Router) {
			r.Delete("/", a.DeleteTaskHandler)
		})
	})
}

func (a *Api) Start() {
	a.InitRouter()
	http.ListenAndServe(fmt.Sprintf("%s:%s", a.Address, a.Port), a.Router)
}
