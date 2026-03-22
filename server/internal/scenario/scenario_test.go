package scenario

import (
	"testing"
)

func TestLoadAll_ReturnsThreeScenarios(t *testing.T) {
	configs, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	const want = 3
	if got := len(configs); got != want {
		t.Errorf("LoadAll() returned %d scenarios, want %d", got, want)
	}
}

func TestLoadAll_ScenarioIDs(t *testing.T) {
	configs, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	expected := []string{"financial_advisor", "brand_marketing", "insurance_claims"}
	for _, id := range expected {
		if _, ok := configs[id]; !ok {
			t.Errorf("LoadAll() missing scenario with id %q", id)
		}
	}
}

func TestLoadAll_FinancialAdvisorFields(t *testing.T) {
	configs, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	cfg, ok := configs["financial_advisor"]
	if !ok {
		t.Fatal("financial_advisor scenario not found")
	}

	if cfg.Name != "financial_advisor" {
		t.Errorf("Name = %q, want %q", cfg.Name, "financial_advisor")
	}
	if cfg.PersonaName != "Morgan" {
		t.Errorf("PersonaName = %q, want %q", cfg.PersonaName, "Morgan")
	}
	if cfg.MaxTurns != 20 {
		t.Errorf("MaxTurns = %d, want 20", cfg.MaxTurns)
	}
	if cfg.TokenBudget != 4000 {
		t.Errorf("TokenBudget = %d, want 4000", cfg.TokenBudget)
	}
	if len(cfg.AllowedIntents) == 0 {
		t.Error("AllowedIntents is empty")
	}
	if len(cfg.BlocklistTerms) == 0 {
		t.Error("BlocklistTerms is empty")
	}
	if len(cfg.OutputConstraints) == 0 {
		t.Error("OutputConstraints is empty")
	}
	if cfg.PersonaDescription == "" {
		t.Error("PersonaDescription is empty")
	}
}

func TestLoadAll_AllScenariosHaveRequiredFields(t *testing.T) {
	configs, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	for id, cfg := range configs {
		if cfg.Name == "" {
			t.Errorf("scenario %q: Name is empty", id)
		}
		if cfg.PersonaName == "" {
			t.Errorf("scenario %q: PersonaName is empty", id)
		}
		if cfg.PersonaDescription == "" {
			t.Errorf("scenario %q: PersonaDescription is empty", id)
		}
		if cfg.MaxTurns == 0 {
			t.Errorf("scenario %q: MaxTurns is zero", id)
		}
		if cfg.TokenBudget == 0 {
			t.Errorf("scenario %q: TokenBudget is zero", id)
		}
	}
}
