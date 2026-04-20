package patterns

import (
	"strings"
	"time"

	"github.com/rkbharti/LogSensei_CLI/internal/config"
	"github.com/rkbharti/LogSensei_CLI/internal/input"
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
// compiled is the pre-compiled pattern slice from config.Config.Compile().
// Pass nil to use only built-in default patterns.
// ─────────────────────────────────────────────────────────────────────────────
func DetectError(parsed input.ParsedLine, lineNum int, context string, compiled []config.CompiledPattern) *ErrorMatch {

	if strings.TrimSpace(parsed.Raw) == "" {
		return nil
	}

	var match *ErrorMatch

	if parsed.IsJSON {
		match = detectFromJSON(parsed, lineNum, context, compiled)
	} else {
		match = detectFromPlainText(parsed.Raw, lineNum, context, compiled)
	}

	if match != nil {
		match.Timestamp = parsed.Timestamp
	}

	return match
}

// ─────────────────────────────────────────────────────────────────────────────
// detectFromJSON
// ─────────────────────────────────────────────────────────────────────────────
func detectFromJSON(parsed input.ParsedLine, lineNum int, context string, compiled []config.CompiledPattern) *ErrorMatch {

	level := parsed.Level

	if level == "info" || level == "debug" || level == "trace" || level == "warn" || level == "warning" {
		return nil
	}

	isErrorLevel := level == "error" || level == "err" ||
		level == "fatal" || level == "critical" || level == "panic"

	if isErrorLevel {
		errType := classifyMessage(parsed.Message, compiled)
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

	return detectFromPlainText(parsed.Message, lineNum, context, compiled)
}

// ─────────────────────────────────────────────────────────────────────────────
// detectFromPlainText
// ─────────────────────────────────────────────────────────────────────────────
func detectFromPlainText(line string, lineNum int, context string, compiled []config.CompiledPattern) *ErrorMatch {

	lower := strings.ToLower(line)
	fields := strings.Fields(line)
	if len(fields) >= 3 {
		levelField := strings.ToLower(fields[2])
		// Skip pure noise levels — only by position, not substring
		if strings.HasPrefix(levelField, "trace") ||
			strings.HasPrefix(levelField, "event") ||
			strings.HasPrefix(levelField, "debug") ||
			strings.HasPrefix(levelField, "info") {
			return nil
		}
	}
	if strings.Contains(lower, "info") || strings.Contains(lower, "debug") {
		return nil
	}

	if name, matched := matchCompiledPattern(line, lower, compiled); matched {
		return &ErrorMatch{
			LineNumber: lineNum,
			Type:       name,
			Message:    line,
			Context:    context,
		}
	}

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
// matchCompiledPattern checks a line against pre-compiled patterns.
// line = original, lower = strings.ToLower(line) — both passed to avoid
// re-computing ToLower inside the hot loop.
// ─────────────────────────────────────────────────────────────────────────────
func matchCompiledPattern(line string, lower string, compiled []config.CompiledPattern) (string, bool) {

	for _, cp := range compiled {

		// ── keyword match (pre-lowercased on both sides) ──────────────────────
		if cp.Keyword != "" {
			if strings.Contains(lower, cp.Keyword) {
				return cp.Name, true
			}
		}

		// ── regex match (pre-compiled — zero allocation) ──────────────────────
		if cp.Regex != nil {
			if cp.Regex.MatchString(line) {
				return cp.Name, true
			}
		}
	}

	return "", false
}

// ─────────────────────────────────────────────────────────────────────────────
// classifyMessage — used by JSON path.
// ─────────────────────────────────────────────────────────────────────────────
func classifyMessage(message string, compiled []config.CompiledPattern) string {
	lower := strings.ToLower(message)

	if name, matched := matchCompiledPattern(message, lower, compiled); matched {
		return name
	}

	return classifyByDefault(lower)
}

// ─────────────────────────────────────────────────────────────────────────────
// classifyByDefault — built-in keyword patterns.
// ─────────────────────────────────────────────────────────────────────────────
func classifyByDefault(lower string) string {

	// ── specific patterns first (order matters — specific before generic) ───

	// Socket / multicast
	if strings.Contains(lower, "setsockopt") ||
		strings.Contains(lower, "mcast") ||
		strings.Contains(lower, "mailslot") {
		return "Socket Error"
	}

	// Protocol / RSVP
	if strings.Contains(lower, "proterr") ||
		strings.Contains(lower, "rsvp") ||
		strings.Contains(lower, "pathtear") {
		return "Protocol Error"
	}

	// EDC (IBM TCP/IP error codes)
	if strings.Contains(lower, "edc") {
		return "EDC Error"
	}

	// Oracle
	if strings.Contains(lower, "ora-") ||
		strings.Contains(lower, "tns:") ||
		strings.Contains(lower, "sp2-") {
		return "Oracle Error"
	}

	// Panic
	if strings.Contains(lower, "panic") {
		return "Panic Error"
	}

	// Fatal
	if strings.Contains(lower, "fatal") {
		return "Fatal Error"
	}

	// Critical
	if strings.Contains(lower, "critical") {
		return "Critical Error"
	}

	// General error — broader match
	if strings.Contains(lower, "error") ||
		strings.Contains(lower, "exception") ||
		strings.Contains(lower, "traceback") ||
		strings.Contains(lower, "stacktrace") {
		return "General Error"
	}

	// Timeout
	if strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "timed out") ||
		strings.Contains(lower, "deadline exceeded") {
		return "Timeout Error"
	}

	// Connection
	if strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "connection reset") ||
		strings.Contains(lower, "no route to host") ||
		strings.Contains(lower, "network unreachable") {
		return "Connection Error"
	}

	// Memory
	if strings.Contains(lower, "out of memory") ||
		strings.Contains(lower, "oom") ||
		strings.Contains(lower, "killed") {
		return "Memory Error"
	}

	// Auth
	if strings.Contains(lower, "unauthorized") ||
		strings.Contains(lower, "forbidden") ||
		strings.Contains(lower, "access denied") ||
		strings.Contains(lower, "permission denied") {
		return "Auth Error"
	}

	// Crash signals
	if strings.Contains(lower, "segfault") ||
		strings.Contains(lower, "sigsegv") ||
		strings.Contains(lower, "sigkill") ||
		strings.Contains(lower, "core dumped") {
		return "Crash Error"
	}

	// Failed operations — last resort generic
	if strings.Contains(lower, "failed") ||
		strings.Contains(lower, "failure") {
		return "Operation Failed"
	}

	return ""
}
