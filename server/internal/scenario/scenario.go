package scenario

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// scenarioFiles embeds all YAML configs from the scenarios subdirectory.
//
//go:embed scenarios/*.yaml
var scenarioFiles embed.FS

// ScenarioConfig holds the parsed configuration for a single scenario.
type ScenarioConfig struct {
	Name               string   `yaml:"name"`
	PersonaName        string   `yaml:"persona_name"`
	PersonaDescription string   `yaml:"persona_description"`
	MaxTurns           int      `yaml:"max_turns"`
	TokenBudget        int      `yaml:"token_budget"`
	AllowedIntents     []string `yaml:"allowed_intents"`
	BlocklistTerms     []string `yaml:"blocklist_terms"`
	OutputConstraints  []string `yaml:"output_constraints"`
}

// LoadAll reads and parses all embedded YAML scenario files.
// The map key is the scenario ID derived from the filename (without extension).
func LoadAll() (map[string]ScenarioConfig, error) {
	configs := make(map[string]ScenarioConfig)

	err := fs.WalkDir(scenarioFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		data, err := scenarioFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		var cfg ScenarioConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		id := strings.TrimSuffix(filepath.Base(path), ".yaml")
		configs[id] = cfg
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load scenarios: %w", err)
	}

	return configs, nil
}
