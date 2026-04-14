package input

import (
	"bufio"
	"os"
)

// process stream file line-by-line

func ProcessFile(filepath string, handle func(string, int)) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	lineNum := 1

	for scanner.Scan() {
		handle(scanner.Text(), lineNum)
		lineNum++
	}
	return scanner.Err()
}
