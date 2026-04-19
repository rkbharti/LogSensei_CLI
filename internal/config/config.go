package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pattern defines one user-defined detection rule from the YAML config.
// A pattern must have either keyword OR regex — not both, not neither.
type Pattern struct {
	Name    string `yaml:"name"`
	Keyword string `yaml:"keyword"` // plain substring match (case-insensitive)
	Regex   string `yaml:"regex"`   // regular expression match
}

// CompiledPattern is a ready-to-use pattern with the regex pre-compiled.
// Created by Config.Compile() — use this in hot paths like DetectError().
type CompiledPattern struct {
	Name    string
	Keyword string         // lowercased for case-insensitive matching
	Regex   *regexp.Regexp // nil if keyword-only pattern
}

// Config holds all patterns loaded from the YAML file.
type Config struct {
	Patterns []Pattern `yaml:"patterns"`
}

// ─────────────────────────────────────────────────────────────────────────────
// LoadConfig reads, parses, validates and compiles patterns from a YAML file.
// ─────────────────────────────────────────────────────────────────────────────
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
// Catches bad regex immediately on startup — not mid-analysis.
// ─────────────────────────────────────────────────────────────────────────────
func (c *Config) ValidatePatterns() error {
	for i, p := range c.Patterns {

		if strings.TrimSpace(p.Name) == "" {
			return fmt.Errorf("pattern #%d: name is required", i+1)
		}

		hasKeyword := strings.TrimSpace(p.Keyword) != ""
		hasRegex := strings.TrimSpace(p.Regex) != ""

		if !hasKeyword && !hasRegex {
			return fmt.Errorf("pattern %q: must have either 'keyword' or 'regex'", p.Name)
		}

		if hasRegex {
			if _, err := regexp.Compile(p.Regex); err != nil {
				return fmt.Errorf("pattern %q: invalid regex %q — %w", p.Name, p.Regex, err)
			}
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Compile returns a []CompiledPattern with all regex pre-compiled.
// Call this once after LoadConfig and pass the result to DetectError.
// Safe to call on a nil Config — returns nil slice.
// ─────────────────────────────────────────────────────────────────────────────
func (c *Config) Compile() []CompiledPattern {
	if c == nil {
		return nil
	}

	compiled := make([]CompiledPattern, 0, len(c.Patterns))

	for _, p := range c.Patterns {
		cp := CompiledPattern{
			Name:    p.Name,
			Keyword: strings.ToLower(strings.TrimSpace(p.Keyword)), // pre-lowercase
		}

		regexStr := strings.TrimSpace(p.Regex)
		if regexStr != "" {
			// ValidatePatterns() already confirmed this compiles —
			// MustCompile is safe here.
			cp.Regex = regexp.MustCompile(regexStr)
		}

		compiled = append(compiled, cp)
	}

	return compiled
}
