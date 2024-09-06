package controllers

import (
	"encoding/json"
	"net/http"
	"os"
)

const JBOX_ANEWS_STATE_FILE_PATH = "jbox_anews_state"

func GetJboxAnewsStateHandler(w http.ResponseWriter, r *http.Request) {
	var state bool
	data, err := os.ReadFile(JBOX_ANEWS_STATE_FILE_PATH)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch string(data) {
	case "0":
		state = false
	case "1":
		state = true
	default:
		http.Error(w, "invalid state", http.StatusInternalServerError)
		return
	}

	jsonResponse, _ := json.Marshal(state)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)

}

func SetJboxAnewsStateHandler(w http.ResponseWriter, r *http.Request) {
	urlParts := splitURLPath(r.URL.Path)
	if len(urlParts) != 3 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	state := urlParts[2]
	if state != "0" && state != "1" {
		http.Error(w, "Bad state", http.StatusBadRequest)
		return
	}

	os.WriteFile(JBOX_ANEWS_STATE_FILE_PATH, []byte(state), 0644)

}
