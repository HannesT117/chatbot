package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"chatbot/server/internal/scenario"
)

// scenarioSummary is the JSON shape returned by GET /api/scenarios.
type scenarioSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PersonaName string `json:"persona_name"`
}

// ScenariosHandler returns an http.HandlerFunc that serves the scenario list.
// configs is the pre-loaded scenario map (keyed by scenario ID).
func ScenariosHandler(configs map[string]scenario.ScenarioConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summaries := make([]scenarioSummary, 0, len(configs))
		for id, cfg := range configs {
			summaries = append(summaries, scenarioSummary{
				ID:          id,
				Name:        cfg.Name,
				PersonaName: cfg.PersonaName,
			})
		}

		sort.Slice(summaries, func(i, j int) bool { return summaries[i].ID < summaries[j].ID })

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summaries) //nolint:errcheck
	}
}
