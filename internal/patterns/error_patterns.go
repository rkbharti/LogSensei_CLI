package patterns

import (
	"strings"

	"github.com/rkbharti/devdebug/internal/config"
)

type ErrorMatch struct {
	LineNumber int
	Type       string
	Message    string
	Context    string
	File       string
}

// DetectError scans a single log line and returns matched error
func DetectError(line string, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	lower := strings.ToLower(line)

	// 🔥 1. CUSTOM CONFIG PATTERNS (SAFE + VALIDATED)
	if cfg != nil {
		for _, p := range cfg.Patterns {

			keyword := strings.TrimSpace(p.Keyword)

			// 🚫 skip invalid patterns
			if keyword == "" {
				continue
			}

			if strings.Contains(lower, strings.ToLower(keyword)) {
				return &ErrorMatch{
					LineNumber: lineNum,
					Type:       p.Name,
					Message:    line,
					Context:    context,
				}
			}
		}
	}

	// 🔥 2. DEFAULT PATTERNS

	if strings.Contains(lower, "panic") {
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       "Panic Error",
			Message:    line,
			Context:    context,
		}
	}

	if strings.Contains(lower, "error") {
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       "General Error",
			Message:    line,
			Context:    context,
		}
	}

	if strings.Contains(lower, "timeout") {
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       "Timeout Error",
			Message:    line,
			Context:    context,
		}
	}

	return nil
}
