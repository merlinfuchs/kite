// file: marketplace.go

package main

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"sync"
	"github.com/google/uuid"
	"log"
)

type Module struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
}

var (
	modules = make(map[string]Module)
	mutex   = sync.Mutex{}
)

func getModules(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var moduleList []Module
	for _, module := range modules {
		moduleList = append(moduleList, module)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(moduleList)
}

func getModuleByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mutex.Lock()
	module, exists := modules[id]
	mutex.Unlock()

	if !exists {
		http.Error(w, "Module not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(module)
}

func addModule(w http.ResponseWriter, r *http.Request) {
	var newModule Module
	if err := json.NewDecoder(r.Body).Decode(&newModule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newModule.ID = uuid.New().String()
	mutex.Lock()
	modules[newModule.ID] = newModule
	mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newModule)
}

func deleteModule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mutex.Lock()
	_, exists := modules[id]
	if exists {
		delete(modules, id)
	}
	mutex.Unlock()

	if !exists {
		http.Error(w, "Module not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/modules", getModules).Methods("GET")
	r.HandleFunc("/modules/{id}", getModuleByID).Methods("GET")
	r.HandleFunc("/modules", addModule).Methods("POST")
	r.HandleFunc("/modules/{id}", deleteModule).Methods("DELETE")

	log.Println("Kite Marketplace API running on port 8080...")
	http.ListenAndServe(":8080", r)
