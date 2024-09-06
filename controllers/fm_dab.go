package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

const STATE_FILE_PATH = "source.json"

type SourceState struct {
	FM  string `json:"fm"`
	DAB string `json:"dab"`
}

func SetFMAndDABHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowedIP(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	source := r.FormValue("source")
	if !isValidFMSource(source) || !isValidDABSource(source) {
		http.Error(w, "Invalid source. FM supports 0, 1, or 2. DAB supports 0, 1, 2, or 3.", http.StatusBadRequest)
		return
	}

	currentSourceState, err := getCurrentFMAndDABSources()
	if err != nil {
		http.Error(w, "Failed to get current source.", http.StatusInternalServerError)
		return
	}

	go zmq_crossfade("fm", currentSourceState.FM, source)
	// go zmq_crossfade("dab", currentSourceState.DAB, source)

	logSourceChange("fm", source)
	// logSourceChange("dab", source)

	newSourceState := SourceState{FM: source, DAB: currentSourceState.DAB}
	saveSourceState(newSourceState)

	fmt.Fprintln(w, "Source set succesfully.")

}

func SetIndividualFMOrDABSourceHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowedIP(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	urlParts := splitURLPath(r.URL.Path)
	if len(urlParts) < 4 {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}

	router := urlParts[2]
	source := urlParts[3]

	if router != "fm" && router != "dab" {
		http.Error(w, "Invalid router. Must be 'fm' or 'dab'.", http.StatusBadRequest)
		return
	}

	if router == "fm" && !isValidFMSource(source) {
		http.Error(w, "Invalid FM source. Must be 0, 1, or 2.", http.StatusBadRequest)
		return
	}

	if router == "dab" && !isValidDABSource(source) {
		http.Error(w, "Invalid DAB source. Must be 0, 1, 2, or 3.", http.StatusBadRequest)
		return
	}

	currentState, err := getCurrentFMAndDABSources()
	if err != nil {
		http.Error(w, "Failed to get current source.", http.StatusInternalServerError)
		return
	}

	var currentSource string
	switch router {
	case "fm":
		currentSource = currentState.FM
		currentState.FM = source

	case "dab":
		currentSource = currentState.DAB
		currentState.DAB = source

	}

	logSourceChange(router, source)
	go zmq_crossfade(router, currentSource, source)
	saveSourceState(currentState)
	fmt.Fprintln(w, "Source set successfully.")
}

// GetJointFMAndDABSourceHandler is for backwards compatability
// It returns the source number if FM and DAB match, otherwise
// returns "s".
func GetJointFMAndDABSourceHandler(w http.ResponseWriter, r *http.Request) {
	currentSourceState, err := getCurrentFMAndDABSources()
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "No source found.", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to read source.", http.StatusInternalServerError)
		}
		return
	}

	if currentSourceState.DAB != currentSourceState.FM {
		fmt.Fprint(w, "s") // sources are split
		return
	}

	fmt.Fprint(w, currentSourceState.DAB) // sources match
}

func GetIndividualFMOrDABSourceHandler(w http.ResponseWriter, r *http.Request) {
	currentSourceState, err := getCurrentFMAndDABSources()
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "No source found.", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to read source.", http.StatusInternalServerError)
		}
		return
	}

	jsonResponse, _ := json.Marshal(currentSourceState)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func isValidDABSource(source string) bool {
	s, err := strconv.Atoi(source)
	if err != nil {
		return false
	}
	return s >= 0 && s <= 3
}

func isValidFMSource(source string) bool {
	s, err := strconv.Atoi(source)
	if err != nil {
		return false
	}
	return s >= 0 && s <= 2
}

func getCurrentFMAndDABSources() (SourceState, error) {
	var state SourceState
	data, err := os.ReadFile(STATE_FILE_PATH)
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}

func saveSourceState(state SourceState) {
	data, _ := json.Marshal(state)
	os.WriteFile(STATE_FILE_PATH, data, 0644)
}

func logSourceChange(sourceType, source string) {
	selLog(fmt.Sprintf("%s selected %s", sourceType, source))
}
