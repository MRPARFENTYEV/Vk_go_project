package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePollHandler(t *testing.T) {
	poll := Poll{
		ID:       "1",
		Question: "Ваш любимый язык?",
		Options:  map[string]int{"Go": 0, "Python": 0, "Java": 0},
	}
	jsonPoll, _ := json.Marshal(poll)

	r := httptest.NewRequest("POST", "/create", bytes.NewReader(jsonPoll))
	w := httptest.NewRecorder()
	createPollHandler(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestVoteHandler(t *testing.T) {
	polls["1"] = &Poll{
		ID:       "1",
		Question: "Ваш любимый язык?",
		Options:  map[string]int{"Go": 0, "Python": 0, "Java": 0},
	}

	vote := VoteRequest{PollID: "1", Option: "Go"}
	jsonVote, _ := json.Marshal(vote)

	r := httptest.NewRequest("POST", "/vote", bytes.NewReader(jsonVote))
	w := httptest.NewRecorder()
	voteHandler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetResultsHandler(t *testing.T) {
	polls["1"] = &Poll{
		ID:       "1",
		Question: "Ваш любимый язык?",
		Options:  map[string]int{"Go": 1, "Python": 0, "Java": 0},
	}

	r := httptest.NewRequest("GET", "/results?id=1", nil)
	w := httptest.NewRecorder()
	getResultsHandler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestClosePollHandler(t *testing.T) {
	polls["1"] = &Poll{
		ID:       "1",
		Question: "Ваш любимый язык?",
		Options:  map[string]int{"Go": 1, "C++": 0, "Java-Script": 0},
	}

	r := httptest.NewRequest("POST", "/close?id=1", nil)
	w := httptest.NewRecorder()
	closePollHandler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
