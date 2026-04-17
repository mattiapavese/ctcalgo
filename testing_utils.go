package ctcalgo

import (
	"bufio"
)

func readFileLines(filePath string) ([]string, error) {
	file, err := assetsFs.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
