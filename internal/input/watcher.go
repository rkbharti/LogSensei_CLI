package input

import (
	"bufio"
	"os"
	"strings"
	"time"
)

// FollowFile watches file without locking it
func FollowFile(filePath string, handle func(string)) error {

	var lastSize int64 = 0

	for {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}

		stat, _ := file.Stat()

		// If file grew
		if stat.Size() > lastSize {

			file.Seek(lastSize, 0)

			scanner := bufio.NewScanner(file)

			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					handle(line)
				}
			}

			lastSize = stat.Size()
		}

		file.Close()

		time.Sleep(1 * time.Second)
	}
}
