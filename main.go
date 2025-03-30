package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/tarantool/go-tarantool"
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
	conn *tarantool.Connection
	mu   sync.Mutex
)

func init() {
	// Инициализация подключения к Tarantool
	var err error
	conn, err = tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
		User: "admin",
		Pass: "admin",
	})
	if err != nil {
		log.Fatalf("Failed to connect to Tarantool: %v", err)
	}

	// Создаем пространство (аналог таблицы) для хранения опросов
	_, err = conn.Eval(`
		box.schema.create_space('polls', {
			if_not_exists = true,
			format = {
				{name = 'id', type = 'string'},
				{name = 'question', type = 'string'},
				{name = 'options', type = 'map'},
				{name = 'closed', type = 'boolean'}
			}
		})
		box.space.polls:create_index('primary', {
			parts = {'id'},
			if_not_exists = true
		})
	`, []interface{}{})
	if err != nil {
		log.Printf("Warning: space creation failed: %v", err)
	}
}

func createPollHandler(w http.ResponseWriter, r *http.Request) {
	var poll Poll
	if err := json.NewDecoder(r.Body).Decode(&poll); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Сохраняем в Tarantool
	_, err := conn.Insert("polls", []interface{}{
		poll.ID,
		poll.Question,
		poll.Options,
		poll.Closed,
	})
	if err != nil {
		http.Error(w, "Failed to save poll", http.StatusInternalServerError)
		return
	}

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

	// Получаем опрос из Tarantool
	resp, err := conn.Select("polls", "primary", 0, 1, tarantool.IterEq, []interface{}{req.PollID})
	if err != nil || len(resp.Data) == 0 {
		http.Error(w, "Poll not found", http.StatusNotFound)
		return
	}

	pollData := resp.Data[0].([]interface{})
	poll := Poll{
		ID:       pollData[0].(string),
		Question: pollData[1].(string),
		Options:  pollData[2].(map[string]int),
		Closed:   pollData[3].(bool),
	}

	if poll.Closed {
		http.Error(w, "Poll is closed", http.StatusForbidden)
		return
	}

	if _, ok := poll.Options[req.Option]; !ok {
		http.Error(w, "Invalid option", http.StatusBadRequest)
		return
	}

	// Обновляем голос в Tarantool
	poll.Options[req.Option]++
	_, err = conn.Update("polls", "primary", []interface{}{req.PollID}, []interface{}{
		[]interface{}{"=", 2, poll.Options}, // Обновляем поле options (индекс 2)
	})
	if err != nil {
		http.Error(w, "Failed to update poll", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(poll)
}

func getResultsHandler(w http.ResponseWriter, r *http.Request) {
	pollID := r.URL.Query().Get("id")

	resp, err := conn.Select("polls", "primary", 0, 1, tarantool.IterEq, []interface{}{pollID})
	if err != nil || len(resp.Data) == 0 {
		http.Error(w, "Poll not found", http.StatusNotFound)
		return
	}

	pollData := resp.Data[0].([]interface{})
	poll := Poll{
		ID:       pollData[0].(string),
		Question: pollData[1].(string),
		Options:  pollData[2].(map[string]int),
		Closed:   pollData[3].(bool),
	}

	json.NewEncoder(w).Encode(poll)
}

func closePollHandler(w http.ResponseWriter, r *http.Request) {
	pollID := r.URL.Query().Get("id")

	mu.Lock()
	defer mu.Unlock()

	// Обновляем статус опроса в Tarantool
	_, err := conn.Update("polls", "primary", []interface{}{pollID}, []interface{}{
		[]interface{}{"=", 3, true}, // Устанавливаем closed=true (индекс 3)
	})
	if err != nil {
		http.Error(w, "Failed to close poll", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "poll closed"})
}

func main() {
	http.HandleFunc("/create", createPollHandler)
	http.HandleFunc("/vote", voteHandler)
	http.HandleFunc("/results", getResultsHandler)
	http.HandleFunc("/close", closePollHandler)

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
