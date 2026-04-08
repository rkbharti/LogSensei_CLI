package stacktrace

import (
	"regexp" //regular expression package(find patterns inside text)
)

type StackInfo struct {
	File string
	Line string
}

// ExtractFileLine extracts file name and line number
func ExtractFileLine(log string) StackInfo {

	// Example match: main.go:45
	re := regexp.MustCompile(`([\w\.]+\.\w+):(\d+)`)
	match := re.FindStringSubmatch(log)

	if len(match) >= 3 {
		return StackInfo{
			File: match[1],
			Line: match[2],
		}
	}

	return StackInfo{
		File: "unknown",
		Line: "unknown",
	}
}
