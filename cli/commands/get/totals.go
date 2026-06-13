package get

import (
	"strings"

	corelog "github.com/yumauri/fbrcm/core/log"
)

func logGetTotals(output string, rows []parameterRow) {
	logger := corelog.For("get")
	logger.Info("total", "output", output, "projects", countOutputProjects(rows), "parameters", countOutputParameters(rows), "values", countOutputValues(rows))
}

func countOutputProjects(rows []parameterRow) int {
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.ProjectID) == "" {
			continue
		}
		seen[row.ProjectID] = struct{}{}
	}
	return len(seen)
}

func countOutputParameters(rows []parameterRow) int {
	total := 0
	for _, row := range rows {
		if strings.TrimSpace(row.Key) == "" {
			continue
		}
		total++
	}
	return total
}

func countOutputValues(rows []parameterRow) int {
	total := 0
	for _, row := range rows {
		total += countValueLines(row.ValueLines)
	}
	return total
}

func countValueLines(lines []valueLine) int {
	total := 0
	for _, line := range lines {
		if line.Missing {
			continue
		}
		total++
	}
	return total
}
