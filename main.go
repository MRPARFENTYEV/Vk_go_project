package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Poll struct {
	ID       string         `json:"id"`
	Question string         `json:"question"`
	Options  map[string]int `json:"options"`
	Closed   bool           `json:"closed"`
}

type VoteRequest struct {
	PollID string `json:"poll_id"`
	Option string `json:"option"`
}

var (
	polls = make(map[string]*Poll)
	mu    sync.Mutex
)

func createPollHandler(w http.ResponseWriter, r *http.Request) {
	var poll Poll
	if err := json.NewDecoder(r.Body).Decode(&poll); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	polls[poll.ID] = &poll
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(poll)
}

func voteHandler(w http.ResponseWriter, r *http.Request) {
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	poll, exists := polls[req.PollID]
	if !exists || poll.Closed {
		http.Error(w, "Poll not found or closed", http.StatusNotFound)
		return
	}

	if _, ok := poll.Options[req.Option]; !ok {
		http.Error(w, "Invalid option", http.StatusBadRequest)
		return
	}

	poll.Options[req.Option]++
	json.NewEncoder(w).Encode(poll)
}

func getResultsHandler(w http.ResponseWriter, r *http.Request) {
	pollID := r.URL.Query().Get("id")

	mu.Lock()
	poll, exists := polls[pollID]
	mu.Unlock()

	if !exists {
		http.Error(w, "Poll not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(poll)
}

func closePollHandler(w http.ResponseWriter, r *http.Request) {
	pollID := r.URL.Query().Get("id")

	mu.Lock()
	defer mu.Unlock()

	poll, exists := polls[pollID]
	if !exists {
		http.Error(w, "Poll not found", http.StatusNotFound)
		return
	}

	poll.Closed = true
	json.NewEncoder(w).Encode(poll)
}

func main() {
	http.HandleFunc("/create", createPollHandler)
	http.HandleFunc("/vote", voteHandler)
	http.HandleFunc("/results", getResultsHandler)
	http.HandleFunc("/close", closePollHandler)

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
