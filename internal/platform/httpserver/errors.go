package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WSError sends an error message over an upgraded WebSocket connection.
func WSError(conn *websocket.Conn, message string) error {
	return conn.WriteJSON(ErrorResponse{
		Message:   message,
		Timestamp: time.Now(),
	})
}

type ErrorResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func ResponseError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	body := ErrorResponse{
		Message:   message,
		Timestamp: time.Now(),
	}
	w.WriteHeader(status)
	writeErrorBody(w, body)
}

func writeErrorBody(w http.ResponseWriter, body ErrorResponse) {
	if err := json.NewEncoder(w).Encode(body); err != nil {
		returnInternalServerError(w)
	}
}

func returnInternalServerError(w http.ResponseWriter) {
	body := ErrorResponse{
		Message:   "Internal server error",
		Timestamp: time.Now(),
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("Error encode 500 error: ", err)
	}
}
