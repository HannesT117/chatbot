package prompt

import (
	"strings"
	"testing"

	"chatbot/server/internal/scenario"
)

func fullConfig() scenario.ScenarioConfig {
	return scenario.ScenarioConfig{
		Name:               "test-scenario",
		PersonaName:        "Aria",
		PersonaDescription: "a helpful financial advisor",
		OutputConstraints:  []string{"Always cite sources", "Use plain language"},
		AllowedIntents:     []string{"portfolio_inquiry", "market_overview"},
		BlocklistTerms:     []string{"competitor", "lawsuit"},
	}
}

func TestBuild_PersonaName(t *testing.T) {
	cfg := fullConfig()
	out := Build(cfg, "tok123")
	if !strings.Contains(out, cfg.PersonaName) {
		t.Errorf("expected PersonaName %q in output, got:\n%s", cfg.PersonaName, out)
	}
}

func TestBuild_PersonaDescription(t *testing.T) {
	cfg := fullConfig()
	out := Build(cfg, "tok123")
	if !strings.Contains(out, cfg.PersonaDescription) {
		t.Errorf("expected PersonaDescription %q in output, got:\n%s", cfg.PersonaDescription, out)
	}
}

func TestBuild_OutputConstraints(t *testing.T) {
	cfg := fullConfig()
	out := Build(cfg, "tok123")
	for _, constraint := range cfg.OutputConstraints {
		if !strings.Contains(out, constraint) {
			t.Errorf("expected OutputConstraint %q in output, got:\n%s", constraint, out)
		}
	}
}

func TestBuild_AllowedIntents(t *testing.T) {
	cfg := fullConfig()
	out := Build(cfg, "tok123")
	for _, intent := range cfg.AllowedIntents {
		if !strings.Contains(out, intent) {
			t.Errorf("expected AllowedIntent %q in output, got:\n%s", intent, out)
		}
	}
}

func TestBuild_BlocklistTerms(t *testing.T) {
	cfg := fullConfig()
	out := Build(cfg, "tok123")
	for _, term := range cfg.BlocklistTerms {
		if !strings.Contains(out, term) {
			t.Errorf("expected BlocklistTerm %q in output, got:\n%s", term, out)
		}
	}
}

func TestBuild_CanaryToken(t *testing.T) {
	cfg := fullConfig()
	canary := "deadbeefcafe1234"
	out := Build(cfg, canary)
	if !strings.Contains(out, canary) {
		t.Errorf("expected canary token %q in output, got:\n%s", canary, out)
	}
}

func TestBuild_EmptySlices(t *testing.T) {
	cfg := scenario.ScenarioConfig{
		PersonaName:        "Bot",
		PersonaDescription: "a simple bot",
	}
	// Must not panic; must produce output with section headers.
	out := Build(cfg, "emptytoken")
	if out == "" {
		t.Error("expected non-empty output for empty slices config")
	}
}

func TestBuild_SectionHeaders(t *testing.T) {
	cfg := fullConfig()
	out := Build(cfg, "tok123")

	headers := []string{
		"## Role",
		"## Constraints",
		"## Allowed Topics",
		"## Do Not Discuss",
		"## Canary",
	}
	for _, h := range headers {
		if !strings.Contains(out, h) {
			t.Errorf("expected section header %q in output, got:\n%s", h, out)
		}
	}
}
