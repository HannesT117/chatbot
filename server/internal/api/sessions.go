package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"chatbot/server/internal/scenario"
	"chatbot/server/internal/session"
)

type createSessionRequest struct {
	ScenarioID string `json:"scenario_id"`
}

type createSessionResponse struct {
	SessionID  string `json:"session_id"`
	ScenarioID string `json:"scenario_id"`
}

// CreateSessionHandler returns an http.HandlerFunc that handles POST /api/sessions.
// It looks up the scenario, creates a new session, saves it to the store, and returns 201.
func CreateSessionHandler(store session.SessionStore, scenarios map[string]scenario.ScenarioConfig, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.ScenarioID == "" {
			http.Error(w, "scenario_id is required", http.StatusBadRequest)
			return
		}

		cfg, ok := scenarios[req.ScenarioID]
		if !ok {
			http.Error(w, "scenario not found", http.StatusNotFound)
			return
		}

		sess, err := session.NewSession(req.ScenarioID, cfg)
		if err != nil {
			http.Error(w, "failed to create session", http.StatusInternalServerError)
			return
		}

		if err := store.Save(sess); err != nil {
			http.Error(w, "failed to save session", http.StatusInternalServerError)
			return
		}

		logger.Info("session created",
			"session_id", sess.ID,
			"scenario_id", sess.ScenarioID,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createSessionResponse{ //nolint:errcheck
			SessionID:  sess.ID,
			ScenarioID: sess.ScenarioID,
		})
	}
}

// DeleteSessionHandler returns an http.HandlerFunc that handles DELETE /api/sessions/{id}.
// It deletes the session from the store (no-op if not found) and returns 204.
func DeleteSessionHandler(store session.SessionStore, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		store.Delete(id) //nolint:errcheck
		logger.Info("session deleted", "session_id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}
