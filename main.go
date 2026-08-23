package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var (
	tasks  = []Task{}
	nextID = 1
	mutex  sync.Mutex
)

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		mutex.Lock()
		json.NewEncoder(w).Encode(tasks)
		mutex.Unlock()

	case http.MethodPost:
		var newTask Task
		err := json.NewDecoder(r.Body).Decode(&newTask)
		if err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		mutex.Lock()
		newTask.ID = nextID
		nextID++
		tasks = append(tasks, newTask)
		mutex.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newTask)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func taskByIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	switch r.Method {
	case http.MethodGet:
		for _, task := range tasks {
			if task.ID == id {
				json.NewEncoder(w).Encode(task)
				return
			}
		}
		http.Error(w, "Task not found", http.StatusNotFound)

	case http.MethodDelete:
		for i, task := range tasks {
			if task.ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.Error(w, "Task not found", http.StatusNotFound)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func main() {
	tasks = append(tasks, Task{ID: 1, Title: "Learn go syntax", Completed: true})
	nextID = 2

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tasks", tasksHandler)
	mux.HandleFunc("POST /api/tasks", tasksHandler)
	mux.HandleFunc("GET /api/tasks/{id}", taskByIdHandler)
	mux.HandleFunc("DELETE /api/tasks/{id}", taskByIdHandler)

	fmt.Println("Go REST API running at http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
