package controllers

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const LOG_FILE_PATH = "sel.log"

func isAllowedIP(_ *http.Request) bool {
	return true
	// TODO: further authentication
	// return strings.Split(r.RemoteAddr, ":")[0] == ALLOWED_IP
}

func selLog(logString string) {
	f, err := os.OpenFile(LOG_FILE_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to log:", err)
		return
	}
	defer f.Close()

	log.SetOutput(f)
	log.Printf("%s: %s\n", time.Now().Format(time.RFC3339), logString)
}

func splitURLPath(path string) []string {
	parts := strings.Split(path, "/")
	var nonEmptyParts []string
	for _, part := range parts {
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}
	return nonEmptyParts
}
