package patterns

import (
	"regexp"
	"strings"
	"time"

	"github.com/rkbharti/devdebug/internal/config"
	"github.com/rkbharti/devdebug/internal/input"
)

type ErrorMatch struct {
	LineNumber int
	Type       string
	Message    string
	Context    string
	File       string
	Timestamp  time.Time
}

// ─────────────────────────────────────────────────────────────────────────────
// DetectError inspects a ParsedLine and returns an ErrorMatch if it is an error.
// ─────────────────────────────────────────────────────────────────────────────
func DetectError(parsed input.ParsedLine, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	if strings.TrimSpace(parsed.Raw) == "" {
		return nil
	}

	var match *ErrorMatch

	if parsed.IsJSON {
		match = detectFromJSON(parsed, lineNum, context, cfg)
	} else {
		match = detectFromPlainText(parsed.Raw, lineNum, context, cfg)
	}

	if match != nil {
		match.Timestamp = parsed.Timestamp
	}

	return match
}

// ─────────────────────────────────────────────────────────────────────────────
// detectFromJSON
// ─────────────────────────────────────────────────────────────────────────────
func detectFromJSON(parsed input.ParsedLine, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	level := parsed.Level

	if level == "info" || level == "debug" || level == "trace" || level == "warn" || level == "warning" {
		return nil
	}

	isErrorLevel := level == "error" || level == "err" ||
		level == "fatal" || level == "critical" || level == "panic"

	if isErrorLevel {
		errType := classifyMessage(parsed.Message, cfg)
		if errType == "" {
			errType = "General Error"
		}
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       errType,
			Message:    parsed.Message,
			Context:    context,
		}
	}

	return detectFromPlainText(parsed.Message, lineNum, context, cfg)
}

// ─────────────────────────────────────────────────────────────────────────────
// detectFromPlainText — keyword + regex matching on a plain text line.
// ─────────────────────────────────────────────────────────────────────────────
func detectFromPlainText(line string, lineNum int, context string, cfg *config.Config) *ErrorMatch {

	lower := strings.ToLower(line)

	// ── noise filter ──────────────────────────────────────────────────────────
	if strings.Contains(lower, "info") || strings.Contains(lower, "debug") {
		return nil
	}

	// ── custom config patterns (keyword + regex) ──────────────────────────────
	if cfg != nil {
		if name, matched := matchConfigPattern(line, cfg); matched {
			return &ErrorMatch{
				LineNumber: lineNum,
				Type:       name,
				Message:    line,
				Context:    context,
			}
		}
	}

	// ── built-in default patterns ─────────────────────────────────────────────
	errType := classifyByDefault(lower)
	if errType == "" {
		return nil
	}

	return &ErrorMatch{
		LineNumber: lineNum,
		Type:       errType,
		Message:    line,
		Context:    context,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// matchConfigPattern checks a line against all user-defined patterns.
// Returns the matched pattern name and true, or ("", false) if no match.
// Keyword match is case-insensitive. Regex match uses the pattern as-is.
// ─────────────────────────────────────────────────────────────────────────────
func matchConfigPattern(line string, cfg *config.Config) (string, bool) {
	lower := strings.ToLower(line)

	for _, p := range cfg.Patterns {

		// ── keyword match ─────────────────────────────────────────────────────
		keyword := strings.TrimSpace(p.Keyword)
		if keyword != "" {
			if strings.Contains(lower, strings.ToLower(keyword)) {
				return p.Name, true
			}
		}

		// ── regex match 🆕 ────────────────────────────────────────────────────
		regexStr := strings.TrimSpace(p.Regex)
		if regexStr != "" {
			// regex is already validated at load time in config.ValidatePatterns()
			// so Compile here will not fail for valid configs.
			// We compile per-match for now — Phase 18 will add a compiled cache.
			re, err := regexp.Compile(regexStr)
			if err != nil {
				continue // defensive — should never happen after validation
			}
			if re.MatchString(line) {
				return p.Name, true
			}
		}
	}

	return "", false
}

// ─────────────────────────────────────────────────────────────────────────────
// classifyMessage — used by JSON path + classifyByDefault.
// Checks config patterns first, then built-in keywords.
// ─────────────────────────────────────────────────────────────────────────────
func classifyMessage(message string, cfg *config.Config) string {

	if cfg != nil {
		if name, matched := matchConfigPattern(message, cfg); matched {
			return name
		}
	}

	return classifyByDefault(strings.ToLower(message))
}

// ─────────────────────────────────────────────────────────────────────────────
// classifyByDefault — built-in keyword patterns.
// Isolated into its own function so it can be tested independently.
// ─────────────────────────────────────────────────────────────────────────────
func classifyByDefault(lower string) string {

	if strings.Contains(lower, "panic") {
		return "Panic Error"
	}
	if strings.Contains(lower, "error:") ||
		strings.HasPrefix(lower, "error ") ||
		strings.Contains(lower, " exception") {
		return "General Error"
	}
	if strings.Contains(lower, "timeout ") ||
		strings.Contains(lower, "request timeout") ||
		strings.Contains(lower, "timed out") ||
		strings.Contains(lower, "connection timeout") {
		return "Timeout Error"
	}

	return ""
}
