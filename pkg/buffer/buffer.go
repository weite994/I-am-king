package buffer

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

func ProcessAsRingBufferToEnd(httpResp *http.Response, maxJobLogLines int) (string, int, *http.Response, error) {
	lines := make([]string, maxJobLogLines)
	validLines := make([]bool, maxJobLogLines)
	totalLines := 0
	writeIndex := 0

	scanner := bufio.NewScanner(httpResp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		totalLines++

		lines[writeIndex] = line
		validLines[writeIndex] = true
		writeIndex = (writeIndex + 1) % maxJobLogLines
	}

	if err := scanner.Err(); err != nil {
		return "", 0, httpResp, fmt.Errorf("failed to read log content: %w", err)
	}

	var result []string
	linesInBuffer := totalLines
	if linesInBuffer > maxJobLogLines {
		linesInBuffer = maxJobLogLines
	}

	startIndex := 0
	if totalLines > maxJobLogLines {
		startIndex = writeIndex
	}

	for i := 0; i < linesInBuffer; i++ {
		idx := (startIndex + i) % maxJobLogLines
		if validLines[idx] {
			result = append(result, lines[idx])
		}
	}

	return strings.Join(result, "\n"), totalLines, httpResp, nil
}
