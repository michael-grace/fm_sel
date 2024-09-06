package controllers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
)

var SELECTOR_ADDRESS = os.Getenv("SELECTOR_ADDRESS")

func callSelector(cmd string) ([]byte, error) {
	// Create a TCP connection to the server
	conn, err := net.Dial("tcp", SELECTOR_ADDRESS)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Receive the initial response (32 bytes)
	buf := make([]byte, 32)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}

	// Send the command
	_, err = conn.Write([]byte(fmt.Sprintf("%s\n", cmd)))
	if err != nil {
		return nil, err
	}

	// Receive the state (16 bytes)
	stateBuf := make([]byte, 16)
	_, err = conn.Read(stateBuf)
	if err != nil {
		return nil, err
	}

	// Close the connection
	conn.Close()
	return stateBuf, nil
}

func GetStudioSelected() (int, error) {
	res, err := callSelector("Q")
	if err != nil {
		return 0, err
	}

	studioString := string(res[0])
	studio, err := strconv.Atoi(studioString)
	return studio, err

}

func SetStudio(studio int) error {
	if studio < 1 || studio > 8 {
		return fmt.Errorf("invalid studio chosen")
	}

	_, err := callSelector(fmt.Sprintf("S%v", studio))
	if err != nil {
		return err
	}

	selLog(fmt.Sprintf("Set selector to Studio %v", studio))
	return nil
}

func GetSelectorLock() (bool, error) {
	res, err := callSelector("Q")
	if err != nil {
		return false, err
	}

	if res[1] == '1' {
		return true, nil
	} else if res[1] == '0' {
		return false, nil
	}

	return false, fmt.Errorf("selector return invalid")
}

func SetSelectorLock(lockState bool) error {
	var err error
	if lockState {
		_, err = callSelector("L")
	} else {
		_, err = callSelector("U")
	}

	if err != nil {
		return err
	}

	selLog(fmt.Sprintf("Set Selector Lock to %v", lockState))
	return nil
}

func SetSelectorHandler(w http.ResponseWriter, r *http.Request) {
	urlParts := splitURLPath(r.URL.Path)

	if len(urlParts) != 3 {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}

	newSourceStr := urlParts[2]
	newSource, err := strconv.Atoi(newSourceStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = SetStudio(newSource); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "OK")

}

func GetSelectorHandler(w http.ResponseWriter, r *http.Request) {
	studio, err := GetStudioSelected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprint(w, studio)
}

func SetSelectorLockHandler(w http.ResponseWriter, r *http.Request) {
	urlParts := splitURLPath(r.URL.Path)

	if len(urlParts) != 3 {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}

	newSelLock := urlParts[2]

	switch newSelLock {
	case "0":
		if err := SetSelectorLock(false); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "1":
		if err := SetSelectorLock(true); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "invalid lock state", http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, "OK")
}

func GetSelectorLockHandler(w http.ResponseWriter, r *http.Request) {
	lock, err := GetSelectorLock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	jsonResponse, _ := json.Marshal(lock)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
