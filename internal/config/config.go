package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pattern defines one user-defined detection rule.
// A pattern must have either keyword OR regex — not both, not neither.
type Pattern struct {
	Name    string `yaml:"name"`
	Keyword string `yaml:"keyword"` // plain substring match (case-insensitive)
	Regex   string `yaml:"regex"`   // 🆕 regular expression match
}

type Config struct {
	Patterns []Pattern `yaml:"patterns"`
}

// LoadConfig reads and parses the YAML config file.
// Returns a validated config or an error if the file cannot be read/parsed.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.ValidatePatterns(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidatePatterns checks every pattern for correctness at load time.
// This catches bad regex patterns immediately on startup — not mid-analysis.
// ─────────────────────────────────────────────────────────────────────────────
func (c *Config) ValidatePatterns() error {
	for i, p := range c.Patterns {

		// name is required
		if strings.TrimSpace(p.Name) == "" {
			return fmt.Errorf("pattern #%d: name is required", i+1)
		}

		hasKeyword := strings.TrimSpace(p.Keyword) != ""
		hasRegex := strings.TrimSpace(p.Regex) != ""

		// must have at least one matcher
		if !hasKeyword && !hasRegex {
			return fmt.Errorf("pattern %q: must have either 'keyword' or 'regex'", p.Name)
		}

		// validate regex compiles correctly — fail fast at load time
		if hasRegex {
			if _, err := regexp.Compile(p.Regex); err != nil {
				return fmt.Errorf("pattern %q: invalid regex %q — %w", p.Name, p.Regex, err)
			}
		}
	}

	return nil
}
