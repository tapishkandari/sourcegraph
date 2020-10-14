package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/apiworker/apiclient"
)

func (h *handler) setupRoutes(router *mux.Router) {
	router.Path("/dequeue").Methods("POST").HandlerFunc(h.handleDequeue)
	router.Path("/setlog").Methods("POST").HandlerFunc(h.handleSetLogContents)
	router.Path("/complete").Methods("POST").HandlerFunc(h.handleComplete)
	router.Path("/heartbeat").Methods("POST").HandlerFunc(h.handleHeartbeat)
}

// POST /dequeue
func (h *handler) handleDequeue(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.DequeueRequest
	if !decodeBody(w, r, &payload) {
		return
	}

	index, dequeued, err := h.Dequeue(r.Context(), payload.IndexerName)
	if err != nil {
		log15.Error("Failed to dequeue index", "err", err)
		http.Error(w, fmt.Sprintf("failed to dequeue index: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if !dequeued {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, index)
}

// POST /setlog
func (h *handler) handleSetLogContents(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.SetLogRequest
	if !decodeBody(w, r, &payload) {
		return
	}

	if err := h.SetLogContents(r.Context(), payload.IndexerName, payload.IndexID, payload.Contents); err != nil {
		log15.Error("Failed to set log contents", "err", err)
		http.Error(w, fmt.Sprintf("failed to set log contents: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /complete
func (h *handler) handleComplete(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.CompleteRequest
	if !decodeBody(w, r, &payload) {
		return
	}

	found, err := h.Complete(r.Context(), payload.IndexerName, payload.IndexID, payload.ErrorMessage)
	if err != nil {
		log15.Error("Failed to complete index job", "err", err)
		http.Error(w, fmt.Sprintf("failed to complete index job: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /heartbeat
func (h *handler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var payload apiclient.HeartbeatRequest
	if !decodeBody(w, r, &payload) {
		return
	}

	if err := h.Heartbeat(r.Context(), payload.IndexerName, payload.IndexIDs); err != nil {
		log15.Error("Failed to acknowledge heartbeat", "err", err)
		http.Error(w, fmt.Sprintf("failed to acknowledge heartbeat: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decodeBody(w http.ResponseWriter, r *http.Request, payload interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal payload: %s", err.Error()), http.StatusBadRequest)
		return false
	}

	return true
}

// copyAll writes the contents of r to w and logs on write failure.
func copyAll(w http.ResponseWriter, r io.Reader) {
	if _, err := io.Copy(w, r); err != nil {
		log15.Error("Failed to write payload to client", "err", err)
	}
}

// writeJSON writes the JSON-encoded payload to w and logs on write failure.
// If there is an encoding error, then a 500-level status is written to w.
func writeJSON(w http.ResponseWriter, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log15.Error("Failed to serialize result", "err", err)
		http.Error(w, fmt.Sprintf("failed to serialize result: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	copyAll(w, bytes.NewReader(data))
}
