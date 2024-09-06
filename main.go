package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/UniversityRadioYork/fm_sel/controllers"
	"github.com/rs/cors"
)

func main() {
	mux := http.NewServeMux()

	// CORS Middleware
	handler := cors.Default().Handler(mux)

	// Set source route for both FM and DAB
	mux.HandleFunc("/source", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SetFMAndDABHandler(w, r)
		} else if r.Method == http.MethodGet {
			controllers.GetJointFMAndDABSourceHandler(w, r)
		} else {
			http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		}
	})

	// Set source route for individual FM or DAB
	mux.HandleFunc("/v2/source/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SetIndividualFMOrDABSourceHandler(w, r)
		} else if r.Method == http.MethodGet {
			controllers.GetIndividualFMOrDABSourceHandler(w, r)
		} else {
			http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v2/studio/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SetSelectorHandler(w, r)
		} else if r.Method == http.MethodGet {
			controllers.GetSelectorHandler(w, r)
		} else {
			http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v2/sellock/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SetSelectorLockHandler(w, r)
		} else if r.Method == http.MethodGet {
			controllers.GetSelectorLockHandler(w, r)
		} else {
			http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v2/programmedata/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SetProgrammeDataStateHandler(w, r)
		} else if r.Method == http.MethodGet {
			controllers.GetProgrammeDataStateHandler(w, r)
		} else {
			http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v2/jboxanews/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SetJboxAnewsStateHandler(w, r)
		} else if r.Method == http.MethodGet {
			controllers.GetJboxAnewsStateHandler(w, r)
		} else {
			http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		}
	})

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	fmt.Println("Server running on port 5001...")
	log.Fatal(http.ListenAndServe(":5001", handler))
}
