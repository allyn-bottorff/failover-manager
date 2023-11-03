// Failover Manager API. Provides a simple random ID at startup.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// RandomID character set.
const CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const PORT = 8080
const IP = "0.0.0.0"

type IDResonse struct {
	ID string `json:"id"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Generate a random ID string.
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = CHARSET[rand.Int63()%int64(len(CHARSET))]
	}
	return string(b)
}

// Readyz health endpoint handler.
func HandleReadyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

// Livez health endpoint handler.
func HandleLivez(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

// Handler for the /ID endpoint. Returns a JSON body containing the random ID.
func HandleID(w http.ResponseWriter, r *http.Request, id IDResonse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(id)
}

func main() {
	//TODO (alb): Load config from json config file
	randomID := IDResonse{
		ID: RandString(24),
	}

	log.Printf("Generating new random ID: %s\n", randomID.ID)

	//Health routes.
	http.HandleFunc("/readyz", HandleReadyz)
	http.HandleFunc("/livez", HandleReadyz)

	//Return ID code generated at startup
	http.HandleFunc("/id", func(w http.ResponseWriter, r *http.Request) {
		HandleID(w, r, randomID)
	})

	log.Printf("Starting HTTPS server. Listening on %d\n", PORT)

	//Start the server

	serverURL := fmt.Sprintf("%s:%d", IP, PORT)
	log.Fatal(http.ListenAndServe(serverURL, nil))

}
