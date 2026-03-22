package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chatbot/server/internal/scenario"
)

func TestScenariosHandler_StatusOK(t *testing.T) {
	configs := map[string]scenario.ScenarioConfig{
		"test_scenario": {Name: "test_scenario", PersonaName: "Tester"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/scenarios", nil)
	w := httptest.NewRecorder()

	ScenariosHandler(configs)(w, req)

	if got := w.Code; got != http.StatusOK {
		t.Errorf("status = %d, want %d", got, http.StatusOK)
	}
}

func TestScenariosHandler_ReturnsSummaries(t *testing.T) {
	configs := map[string]scenario.ScenarioConfig{
		"financial_advisor": {Name: "financial_advisor", PersonaName: "Morgan"},
		"brand_marketing":   {Name: "brand_marketing", PersonaName: "Sage"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/scenarios", nil)
	w := httptest.NewRecorder()

	ScenariosHandler(configs)(w, req)

	var summaries []scenarioSummary
	if err := json.NewDecoder(w.Body).Decode(&summaries); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got := len(summaries); got != 2 {
		t.Errorf("len(summaries) = %d, want 2", got)
	}
}

func TestScenariosHandler_SummaryFields(t *testing.T) {
	configs := map[string]scenario.ScenarioConfig{
		"financial_advisor": {Name: "financial_advisor", PersonaName: "Morgan"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/scenarios", nil)
	w := httptest.NewRecorder()

	ScenariosHandler(configs)(w, req)

	var summaries []scenarioSummary
	if err := json.NewDecoder(w.Body).Decode(&summaries); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	s := summaries[0]
	if s.ID != "financial_advisor" {
		t.Errorf("ID = %q, want %q", s.ID, "financial_advisor")
	}
	if s.Name != "financial_advisor" {
		t.Errorf("Name = %q, want %q", s.Name, "financial_advisor")
	}
	if s.PersonaName != "Morgan" {
		t.Errorf("PersonaName = %q, want %q", s.PersonaName, "Morgan")
	}
}

func TestScenariosHandler_EmptyConfigs(t *testing.T) {
	configs := map[string]scenario.ScenarioConfig{}

	req := httptest.NewRequest(http.MethodGet, "/api/scenarios", nil)
	w := httptest.NewRecorder()

	ScenariosHandler(configs)(w, req)

	if got := w.Code; got != http.StatusOK {
		t.Errorf("status = %d, want %d", got, http.StatusOK)
	}

	var summaries []scenarioSummary
	if err := json.NewDecoder(w.Body).Decode(&summaries); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(summaries) != 0 {
		t.Errorf("expected empty summaries, got %d", len(summaries))
	}
}
