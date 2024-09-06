package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type RunningState int

const (
	StateRun RunningState = iota
	StatePause
	StateStop
)

var runningStates []RunningState = []RunningState{StateRun, StatePause, StateStop}

const PROGRAMME_DATA_STATE_FILE_PATH = "programmeDataState.json"

type ProgrammeDataRunningState struct {
	FM  RunningState `json:"fm"`
	DAB RunningState `json:"dab"`
}

func GetProgrammeDataStateHandler(w http.ResponseWriter, r *http.Request) {
	currentState, err := getCurrentProgrammeDataState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse, _ := json.Marshal(currentState)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func SetProgrammeDataStateHandler(w http.ResponseWriter, r *http.Request) {
	currentState, err := getCurrentProgrammeDataState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	urlParts := splitURLPath(r.URL.Path)
	if len(urlParts) != 4 {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}

	output := urlParts[2]
	stateStr := urlParts[3]

	state, err := strconv.Atoi(stateStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	switch output {
	case "fm":
		currentState.FM = runningStates[state]
	case "dab":
		currentState.DAB = runningStates[state]
	default:
		http.Error(w, "Invalid output", http.StatusBadRequest)
		return
	}

	selLog(fmt.Sprintf("%s programme data set to %v", output, state))

	data, _ := json.Marshal(currentState)
	os.WriteFile(PROGRAMME_DATA_STATE_FILE_PATH, data, 0644)

	fmt.Fprintln(w, "State set successfully")

}

func getCurrentProgrammeDataState() (ProgrammeDataRunningState, error) {
	var state ProgrammeDataRunningState
	data, err := os.ReadFile(PROGRAMME_DATA_STATE_FILE_PATH)
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}
