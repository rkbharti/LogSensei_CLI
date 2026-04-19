package input

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
)

// FollowFile watches a file for new lines using OS-level file events.
// It calls handle() for every new non-empty line written to the file.
// Blocks until an error occurs or the watcher is closed.
func FollowFile(filePath string, handle func(string)) error {

	// ── open the file and seek to end (only watch new lines) ─────────────────
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// seek to end — we only care about NEW lines written after watch starts
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("cannot seek file: %w", err)
	}

	reader := bufio.NewReader(file)

	// ── create fsnotify watcher ───────────────────────────────────────────────
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("cannot create watcher: %w", err)
	}
	defer watcher.Close()

	// watch the file
	if err := watcher.Add(filePath); err != nil {
		return fmt.Errorf("cannot watch file: %w", err)
	}

	// ── event loop ────────────────────────────────────────────────────────────
	for {
		select {

		case event, ok := <-watcher.Events:
			if !ok {
				// channel closed — watcher shut down
				return nil
			}

			// only react to write events (new log lines appended)
			if event.Has(fsnotify.Write) {
				if err := readNewLines(reader, handle); err != nil {
					return err
				}
			}

			// file was removed or renamed (log rotation)
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				return fmt.Errorf("file was removed or rotated: %s", filePath)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// readNewLines reads all lines added since the last read position.
// Called on every Write event — drains only the new bytes.
// ─────────────────────────────────────────────────────────────────────────────

func readNewLines(reader *bufio.Reader, handle func(string)) error {
	for {
		line, err := reader.ReadString('\n')

		// even if err != nil, line may contain partial data — process it first
		if len(line) > 0 {
			trimmed := trimLine(line)
			if trimmed != "" {
				handle(trimmed)
			}
		}

		if err != nil {
			// io.EOF is normal — no more new data right now, stop reading
			// any other error is a real problem
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}
	}
}

// trimLine removes trailing newline characters and whitespace.
func trimLine(line string) string {
	// trim \r\n (Windows) and \n (Unix)
	result := line
	for len(result) > 0 && (result[len(result)-1] == '\n' || result[len(result)-1] == '\r') {
		result = result[:len(result)-1]
	}
	return result
}
