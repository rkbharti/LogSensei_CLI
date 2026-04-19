package input

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// helper
// ─────────────────────────────────────────────────────────────────────────────

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp log: %v", err)
	}
	return path
}

// ─────────────────────────────────────────────────────────────────────────────
// ParseLine — plain text
// ─────────────────────────────────────────────────────────────────────────────

func TestParseLine_PlainText(t *testing.T) {

	tests := []struct {
		name        string
		input       string
		wantRaw     string
		wantMessage string
		wantIsJSON  bool
		wantLevel   string
	}{
		{
			name:        "plain error line",
			input:       "error: failed to open config",
			wantRaw:     "error: failed to open config",
			wantMessage: "error: failed to open config",
			wantIsJSON:  false,
			wantLevel:   "",
		},
		{
			name:        "empty string",
			input:       "",
			wantRaw:     "",
			wantMessage: "",
			wantIsJSON:  false,
			wantLevel:   "",
		},
		{
			name:        "whitespace-only line",
			input:       "   ",
			wantRaw:     "   ",
			wantMessage: "   ",
			wantIsJSON:  false,
			wantLevel:   "",
		},
		{
			name:        "INFO log line",
			input:       "[INFO] Server started on port 8080",
			wantRaw:     "[INFO] Server started on port 8080",
			wantMessage: "[INFO] Server started on port 8080",
			wantIsJSON:  false,
			wantLevel:   "",
		},
		{
			name:        "panic line",
			input:       "panic: runtime error: nil pointer dereference",
			wantRaw:     "panic: runtime error: nil pointer dereference",
			wantMessage: "panic: runtime error: nil pointer dereference",
			wantIsJSON:  false,
			wantLevel:   "",
		},
		{
			name:        "broken JSON treated as plain text",
			input:       `{"level":"error", broken`,
			wantRaw:     `{"level":"error", broken`,
			wantMessage: `{"level":"error", broken`,
			wantIsJSON:  false,
			wantLevel:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLine(tt.input)

			if result.Raw != tt.wantRaw {
				t.Errorf("Raw: got %q, want %q", result.Raw, tt.wantRaw)
			}
			if result.Message != tt.wantMessage {
				t.Errorf("Message: got %q, want %q", result.Message, tt.wantMessage)
			}
			if result.IsJSON != tt.wantIsJSON {
				t.Errorf("IsJSON: got %v, want %v", result.IsJSON, tt.wantIsJSON)
			}
			if result.Level != tt.wantLevel {
				t.Errorf("Level: got %q, want %q", result.Level, tt.wantLevel)
			}
			if !result.Timestamp.IsZero() {
				t.Errorf("Timestamp: expected zero for plain text, got %v", result.Timestamp)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ParseLine — JSON
// ─────────────────────────────────────────────────────────────────────────────

func TestParseLine_JSON(t *testing.T) {

	tests := []struct {
		name        string
		input       string
		wantMessage string
		wantLevel   string
		wantIsJSON  bool
	}{
		{
			name:        "JSON with message field",
			input:       `{"level":"error","message":"database connection failed"}`,
			wantMessage: "database connection failed",
			wantLevel:   "error",
			wantIsJSON:  true,
		},
		{
			name:        "JSON with msg field (alternative)",
			input:       `{"level":"fatal","msg":"server crashed"}`,
			wantMessage: "server crashed",
			wantLevel:   "fatal",
			wantIsJSON:  true,
		},
		{
			name:        "JSON level is lowercased",
			input:       `{"level":"ERROR","message":"something failed"}`,
			wantMessage: "something failed",
			wantLevel:   "error",
			wantIsJSON:  true,
		},
		{
			name:        "JSON info level extracted",
			input:       `{"level":"info","message":"Server started"}`,
			wantMessage: "Server started",
			wantLevel:   "info",
			wantIsJSON:  true,
		},
		{
			name:        "JSON with severity field instead of level",
			input:       `{"severity":"CRITICAL","message":"disk full"}`,
			wantMessage: "disk full",
			wantLevel:   "critical",
			wantIsJSON:  true,
		},
		{
			name:        "JSON with no message key falls back to raw",
			input:       `{"level":"error","service":"api","code":500}`,
			wantMessage: `{"level":"error","service":"api","code":500}`,
			wantLevel:   "error",
			wantIsJSON:  true,
		},
		{
			name:        "JSON message key takes priority over msg",
			input:       `{"message":"primary message","msg":"secondary"}`,
			wantMessage: "primary message", // message checked before msg
			wantIsJSON:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLine(tt.input)

			if result.IsJSON != tt.wantIsJSON {
				t.Errorf("IsJSON: got %v, want %v", result.IsJSON, tt.wantIsJSON)
			}
			if result.Message != tt.wantMessage {
				t.Errorf("Message: got %q, want %q", result.Message, tt.wantMessage)
			}
			if result.Level != tt.wantLevel {
				t.Errorf("Level: got %q, want %q", result.Level, tt.wantLevel)
			}
			if result.Raw != tt.input {
				t.Errorf("Raw: got %q, want %q (original)", result.Raw, tt.input)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ParseLine — timestamp extraction
// ─────────────────────────────────────────────────────────────────────────────

func TestParseLine_Timestamp(t *testing.T) {

	t.Run("RFC3339 timestamp extracted", func(t *testing.T) {
		line := `{"level":"error","message":"crash","time":"2026-04-19T14:00:00Z"}`
		result := ParseLine(line)

		if result.Timestamp.IsZero() {
			t.Fatal("expected non-zero timestamp, got zero")
		}
		want := time.Date(2026, 4, 19, 14, 0, 0, 0, time.UTC)
		if !result.Timestamp.Equal(want) {
			t.Errorf("Timestamp: got %v, want %v", result.Timestamp, want)
		}
	})

	t.Run("ts field as Unix seconds extracted", func(t *testing.T) {
		line := `{"level":"error","msg":"crash","ts":1713513600}`
		result := ParseLine(line)

		if result.Timestamp.IsZero() {
			t.Fatal("expected non-zero timestamp, got zero")
		}
		want := time.Unix(1713513600, 0).UTC()
		if !result.Timestamp.Equal(want) {
			t.Errorf("Timestamp: got %v, want %v", result.Timestamp, want)
		}
	})

	t.Run("ts field as Unix milliseconds extracted", func(t *testing.T) {
		line := `{"level":"error","msg":"crash","ts":1713513600000}`
		result := ParseLine(line)

		if result.Timestamp.IsZero() {
			t.Fatal("expected non-zero timestamp, got zero")
		}
		want := time.UnixMilli(1713513600000).UTC()
		if !result.Timestamp.Equal(want) {
			t.Errorf("Timestamp: got %v, want %v", result.Timestamp, want)
		}
	})

	t.Run("JSON with no timestamp field has zero Timestamp", func(t *testing.T) {
		line := `{"level":"error","message":"crash"}`
		result := ParseLine(line)

		if !result.Timestamp.IsZero() {
			t.Errorf("expected zero Timestamp, got %v", result.Timestamp)
		}
	})

	t.Run("plain text has zero Timestamp", func(t *testing.T) {
		result := ParseLine("panic: nil pointer dereference")
		if !result.Timestamp.IsZero() {
			t.Errorf("expected zero Timestamp for plain text, got %v", result.Timestamp)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// ParseLine — Fields map
// ─────────────────────────────────────────────────────────────────────────────

func TestParseLine_Fields(t *testing.T) {

	t.Run("string fields collected in Fields map", func(t *testing.T) {
		line := `{"level":"error","message":"crash","service":"api","host":"prod-01"}`
		result := ParseLine(line)

		if result.Fields == nil {
			t.Fatal("Fields map should not be nil for JSON line")
		}
		if result.Fields["service"] != "api" {
			t.Errorf("Fields[service]: got %q, want %q", result.Fields["service"], "api")
		}
		if result.Fields["host"] != "prod-01" {
			t.Errorf("Fields[host]: got %q, want %q", result.Fields["host"], "prod-01")
		}
	})

	t.Run("plain text has nil Fields map", func(t *testing.T) {
		result := ParseLine("error: something went wrong")
		if result.Fields != nil {
			t.Errorf("Fields should be nil for plain text, got %v", result.Fields)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// ProcessFile
// ─────────────────────────────────────────────────────────────────────────────

func TestProcessFile(t *testing.T) {

	t.Run("reads all lines and increments line numbers", func(t *testing.T) {
		path := writeTempLog(t, "line one\nline two\nline three\n")

		var lines []string
		var nums []int

		err := ProcessFile(path, func(parsed ParsedLine, lineNum int) {
			lines = append(lines, parsed.Raw)
			nums = append(nums, lineNum)
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(lines))
		}
		if lines[0] != "line one" {
			t.Errorf("line 1: got %q, want %q", lines[0], "line one")
		}
		if nums[0] != 1 || nums[1] != 2 || nums[2] != 3 {
			t.Errorf("line numbers: got %v, want [1 2 3]", nums)
		}
	})

	t.Run("empty file calls handler zero times", func(t *testing.T) {
		path := writeTempLog(t, "")
		count := 0
		err := ProcessFile(path, func(_ ParsedLine, _ int) { count++ })

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 calls, got %d", count)
		}
	})

	t.Run("JSON lines are parsed correctly by ProcessFile", func(t *testing.T) {
		content := `{"level":"error","message":"database failed"}` + "\n" +
			`{"level":"info","message":"Server started"}` + "\n"

		path := writeTempLog(t, content)

		var results []ParsedLine
		ProcessFile(path, func(parsed ParsedLine, _ int) {
			results = append(results, parsed)
		})

		if len(results) != 2 {
			t.Fatalf("expected 2 parsed lines, got %d", len(results))
		}
		if !results[0].IsJSON {
			t.Error("first line should be JSON")
		}
		if results[0].Message != "database failed" {
			t.Errorf("first message: got %q, want %q", results[0].Message, "database failed")
		}
		if results[0].Level != "error" {
			t.Errorf("first level: got %q, want %q", results[0].Level, "error")
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		err := ProcessFile("nonexistent/path/file.log", func(_ ParsedLine, _ int) {})
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})

	t.Run("mixed plain and JSON lines processed correctly", func(t *testing.T) {
		content := "panic: nil pointer dereference\n" +
			`{"level":"error","message":"timeout"}` + "\n" +
			"[INFO] Server started\n"

		path := writeTempLog(t, content)

		var results []ParsedLine
		ProcessFile(path, func(parsed ParsedLine, _ int) {
			results = append(results, parsed)
		})

		if len(results) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(results))
		}
		if results[0].IsJSON {
			t.Error("line 1 (plain) should not be JSON")
		}
		if !results[1].IsJSON {
			t.Error("line 2 (JSON) should be IsJSON=true")
		}
		if results[2].IsJSON {
			t.Error("line 3 (plain) should not be JSON")
		}
	})
}
