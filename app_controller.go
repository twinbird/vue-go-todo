package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func appGetHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, AppPageTemplate, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

type TaskResponse struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	State     int       `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

type TasksResponse struct {
	OwnerId int            `json:"owner_id"`
	Tasks   []TaskResponse `json:"tasks"`
}

func taskGetHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value(ContextUserIdKey).(int)

	q := r.FormValue("q")
	state, err := strconv.Atoi(r.FormValue("state"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks, err := getTasks(user_id, state, q)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := &TasksResponse{OwnerId: user_id, Tasks: make([]TaskResponse, 0)}
	for _, task := range tasks {
		t := TaskResponse{task.Id, task.Title, task.State, task.CreatedAt}
		res.Tasks = append(res.Tasks, t)
	}
	b, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

type CreateResponse struct {
	OwnerId     int          `json:"owner_id"`
	CreatedTask TaskResponse `json:"created_task"`
}

func taskPostHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value(ContextUserIdKey).(int)

	title := r.FormValue("title")
	if len(title) > 100 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "titleは100文字以内にしてください。", http.StatusBadRequest)
		return
	}
	t := &Task{Title: title, UserId: user_id, CreatedAt: time.Now()}
	err := t.Create()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	restask := TaskResponse{t.Id, t.Title, t.State, t.CreatedAt}
	res := &CreateResponse{
		OwnerId:     user_id,
		CreatedTask: restask,
	}
	b, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

type UpdateResponse struct {
	OwnerId     int          `json:"owner_id"`
	UpdatedTask TaskResponse `json:"updated_task"`
}

func taskPutHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value(ContextUserIdKey).(int)

	params := mux.Vars(r)
	task_id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	title := r.FormValue("title")
	state, err := strconv.Atoi(r.FormValue("state"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t := &Task{Id: task_id, Title: title, UserId: user_id, State: state}
	err = t.Update()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	restask := TaskResponse{t.Id, t.Title, t.State, t.CreatedAt}
	res := &UpdateResponse{
		OwnerId:     user_id,
		UpdatedTask: restask,
	}
	b, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

type RemoveResponse struct {
	OwnerId       int   `json:"owner_id"`
	RemovedTaskId []int `json:"removed_task_id"`
}

func taskDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value(ContextUserIdKey).(int)

	ids, err := removeDoneTask(user_id)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := &RemoveResponse{OwnerId: user_id, RemovedTaskId: ids}
	b, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
