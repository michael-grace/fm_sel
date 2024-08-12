package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	zmq "github.com/pebbe/zmq4"
	"github.com/rs/cors"
)

const (
	FILE_PATH  = "source.json"
	LOG_FILE   = "sel.log"
	ALLOWED_IP = "127.0.0.1"
)

type SourceConfig struct {
	FM  string `json:"fm"`
	DAB string `json:"dab"`
}

const DO_ZMQ_CROSSFADES = true

func main() {
	mux := http.NewServeMux()

	// CORS Middleware
	handler := cors.Default().Handler(mux)

	// Set source route for both FM and DAB
	mux.HandleFunc("/source", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			setBothSources(w, r)
		} else if r.Method == http.MethodGet {
			getJointSource(w, r)
		} else {
			http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		}
	})

	// Set source route for individual FM or DAB
	mux.HandleFunc("/v2/source/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			setIndividualSource(w, r)
		} else if r.Method == http.MethodGet {
			getSource(w, r)
		} else {
			http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		}
	})

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	fmt.Println("Server running on port 5001...")
	log.Fatal(http.ListenAndServe(":5001", handler))
}

func setBothSources(w http.ResponseWriter, r *http.Request) {
	if !isAllowedIP(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	source := r.FormValue("source")
	if !isValidFMSource(source) || !isValidDABSource(source) {
		http.Error(w, "Invalid source. FM supports 0, 1, or 2. DAB supports 0, 1, 2, or 3.", http.StatusBadRequest)
		return
	}

	currentConfig, err := getCurrentSourceConfig()
	if err != nil {
		http.Error(w, "Failed to get current source.", http.StatusInternalServerError)
		return
	}

	logSourceSelection("fm", source)
	logSourceSelection("dab", source)

	if DO_ZMQ_CROSSFADES {
		go crossfade("fm", currentConfig.FM, source)
		crossfade("dab", currentConfig.DAB, source)
	}

	newConfig := SourceConfig{FM: source, DAB: source}
	saveSourceConfig(newConfig)

	fmt.Fprintln(w, "Source set successfully.")
}

func setIndividualSource(w http.ResponseWriter, r *http.Request) {
	if !isAllowedIP(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse the URL parameters
	parts := splitPath(r.URL.Path)
	if len(parts) < 4 {
		http.Error(w, "Invalid path.", http.StatusBadRequest)
		return
	}

	sourceType := parts[2]
	source := parts[3]

	if sourceType != "fm" && sourceType != "dab" {
		http.Error(w, "Invalid source type. Must be 'fm' or 'dab'.", http.StatusBadRequest)
		return
	}

	if sourceType == "fm" && !isValidFMSource(source) {
		http.Error(w, "Invalid FM source. Must be 0, 1, or 2.", http.StatusBadRequest)
		return
	}

	if sourceType == "dab" && !isValidDABSource(source) {
		http.Error(w, "Invalid DAB source. Must be 0, 1, 2, or 3.", http.StatusBadRequest)
		return
	}

	currentConfig, err := getCurrentSourceConfig()
	if err != nil {
		http.Error(w, "Failed to get current source.", http.StatusInternalServerError)
		return
	}

	currentSource := ""
	if sourceType == "fm" {
		currentSource = currentConfig.FM
	} else if sourceType == "dab" {
		currentSource = currentConfig.DAB
	}

	logSourceSelection(sourceType, source)
	if DO_ZMQ_CROSSFADES {
		crossfade(sourceType, currentSource, source)
	}

	if sourceType == "fm" {
		currentConfig.FM = source
	} else {
		currentConfig.DAB = source
	}

	saveSourceConfig(currentConfig)

	fmt.Fprintln(w, "Source set successfully.")
}

func getJointSource(w http.ResponseWriter, _ *http.Request) {
	sourceConfig, err := getCurrentSourceConfig()
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "No source found.", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to read source.", http.StatusInternalServerError)
		}
		return
	}

	if sourceConfig.DAB != sourceConfig.FM {
		fmt.Fprint(w, "s") // split source
		return
	}

	fmt.Fprint(w, sourceConfig.FM) // same source
}

func getSource(w http.ResponseWriter, _ *http.Request) {
	sourceConfig, err := getCurrentSourceConfig()
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "No source found.", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to read source.", http.StatusInternalServerError)
		}
		return
	}

	jsonResponse, _ := json.Marshal(sourceConfig)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func getCurrentSourceConfig() (SourceConfig, error) {
	var config SourceConfig
	data, err := os.ReadFile(FILE_PATH)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func saveSourceConfig(config SourceConfig) {
	data, _ := json.Marshal(config)
	os.WriteFile(FILE_PATH, data, 0644)
}

func logSourceSelection(sourceType, source string) {
	f, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to log source selection:", err)
		return
	}
	defer f.Close()

	log.SetOutput(f)
	log.Printf("%s: %s selected %s\n", time.Now().Format(time.RFC3339), sourceType, source)
}

func crossfade(sourceType, currentSource, newSource string) {
	socket, _ := zmq.NewSocket(zmq.REQ)
	defer socket.Close()

	address := "tcp://localhost:5555"
	if sourceType == "dab" {
		address = "tcp://localhost:5556"
	}

	socket.Connect(address)

	for i := 1; i <= 5; i++ {
		fadeVolume(socket, currentSource, 5-i)
		fadeVolume(socket, newSource, i)

		time.Sleep(600 * time.Millisecond)
	}
}

func fadeVolume(socket *zmq.Socket, source string, level int) {
	msg := fmt.Sprintf("volume@s%s volume %.1f", source, float64(level)/5)
	socket.Send(msg, 0)
	socket.Recv(0)
}

func isValidFMSource(source string) bool {
	return source == "0" || source == "1" || source == "2"
}

func isValidDABSource(source string) bool {
	return source == "0" || source == "1" || source == "2" || source == "3"
}

func isAllowedIP(r *http.Request) bool {
	return true
	// TODO: further authentication
	return strings.Split(r.RemoteAddr, ":")[0] == ALLOWED_IP
}

func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	var nonEmptyParts []string
	for _, part := range parts {
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}
	return nonEmptyParts
}
