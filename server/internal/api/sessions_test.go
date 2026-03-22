package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"chatbot/server/internal/api"
	"chatbot/server/internal/scenario"
	"chatbot/server/internal/session"
)

// mockStore is a minimal SessionStore for testing.
type mockStore struct {
	sessions map[string]*session.Session
	saveErr  error
}

func newMockStore() *mockStore {
	return &mockStore{sessions: make(map[string]*session.Session)}
}

func (m *mockStore) Get(id string) (*session.Session, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return s, nil
}

func (m *mockStore) Save(s *session.Session) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.sessions[s.ID] = s
	return nil
}

func (m *mockStore) Delete(id string) error {
	delete(m.sessions, id)
	return nil
}

var testScenarios = map[string]scenario.ScenarioConfig{
	"fin_advisor": {
		Name:        "Financial Advisor",
		PersonaName: "Alex",
		MaxTurns:    10,
		TokenBudget: 2000,
	},
}

var silentLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestCreateSessionHandler_ValidScenario(t *testing.T) {
	store := newMockStore()
	h := api.CreateSessionHandler(store, testScenarios, silentLogger)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions",
		bytes.NewBufferString(`{"scenario_id":"fin_advisor"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var resp struct {
		SessionID  string `json:"session_id"`
		ScenarioID string `json:"scenario_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.SessionID == "" {
		t.Error("expected non-empty session_id")
	}
	if resp.ScenarioID != "fin_advisor" {
		t.Errorf("expected scenario_id fin_advisor, got %s", resp.ScenarioID)
	}
}

func TestCreateSessionHandler_MissingScenarioID(t *testing.T) {
	store := newMockStore()
	h := api.CreateSessionHandler(store, testScenarios, silentLogger)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions",
		bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateSessionHandler_UnknownScenario(t *testing.T) {
	store := newMockStore()
	h := api.CreateSessionHandler(store, testScenarios, silentLogger)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions",
		bytes.NewBufferString(`{"scenario_id":"unknown"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateSessionHandler_InvalidBody(t *testing.T) {
	store := newMockStore()
	h := api.CreateSessionHandler(store, testScenarios, silentLogger)

	req := httptest.NewRequest(http.MethodPost, "/api/sessions",
		bytes.NewBufferString(`not json`))
	w := httptest.NewRecorder()

	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteSessionHandler(t *testing.T) {
	store := newMockStore()
	h := api.DeleteSessionHandler(store, silentLogger)

	// Use a mux so PathValue("id") is populated.
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /api/sessions/{id}", h)

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/test123", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
