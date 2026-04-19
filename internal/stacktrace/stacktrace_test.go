package stacktrace

import (
	"testing"
)

func TestExtractFileLine(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		wantFile string
		wantLine string
	}{
		{
			name:     "simple file:line pattern",
			input:    "main.go:45",
			wantFile: "main.go",
			wantLine: "45",
		},
		{
			name:     "file:line inside a longer message",
			input:    "panic: runtime error at server.go:120",
			wantFile: "server.go",
			wantLine: "120",
		},
		{
			name:     "Go stack trace line",
			input:    "goroutine 1 [running]: main.handleRequest(0xc000014080) /home/user/app/handler.go:88",
			wantFile: "handler.go",
			wantLine: "88",
		},
		{
			name:     "underscored file name",
			input:    "error in file_reader.go:33",
			wantFile: "file_reader.go",
			wantLine: "33",
		},
		{
			name:     "file with path — last segment matched",
			input:    "error at /app/internal/config/config.go:72",
			wantFile: "config.go",
			wantLine: "72",
		},
		{
			name:     "no file:line returns unknown for both",
			input:    "something went wrong but no file reference",
			wantFile: "unknown",
			wantLine: "unknown",
		},
		{
			name:     "empty string returns unknown",
			input:    "",
			wantFile: "unknown",
			wantLine: "unknown",
		},
		{
			name:     "whitespace-only returns unknown",
			input:    "   ",
			wantFile: "unknown",
			wantLine: "unknown",
		},
		{
			name:     "colon with no number returns unknown",
			input:    "something: no line number here",
			wantFile: "unknown",
			wantLine: "unknown",
		},
		{
			name:     "multiple file:line — first one matched",
			input:    "error at api.go:10 called from router.go:55",
			wantFile: "api.go",
			wantLine: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFileLine(tt.input)

			if result.File != tt.wantFile {
				t.Errorf("File: got %q, want %q (input: %q)", result.File, tt.wantFile, tt.input)
			}
			if result.Line != tt.wantLine {
				t.Errorf("Line: got %q, want %q (input: %q)", result.Line, tt.wantLine, tt.input)
			}
		})
	}
}

func TestStackInfo_Fields(t *testing.T) {
	t.Run("StackInfo has correct field names", func(t *testing.T) {
		s := StackInfo{File: "main.go", Line: "42"}
		if s.File != "main.go" {
			t.Errorf("File field: got %q, want %q", s.File, "main.go")
		}
		if s.Line != "42" {
			t.Errorf("Line field: got %q, want %q", s.Line, "42")
		}
	})
}
