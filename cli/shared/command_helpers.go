package shared

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/yumauri/fbrcm/core"
)

// HasFilters reports whether any filter query is non-empty after parsing.
func HasFilters(rawFilters []string) bool {
	return len(ParseFilters(rawFilters)) > 0
}

// ResolveParameterArgFilters turns an optional parameter argument into an exact filter.
func ResolveParameterArgFilters(args []string, rawFilters []string) ([]string, error) {
	if len(args) == 0 {
		return rawFilters, nil
	}
	if HasFilters(rawFilters) {
		return nil, fmt.Errorf("parameter argument cannot be used together with --filter")
	}
	return []string{"=" + args[0]}, nil
}

// SortProjects orders projects the same way across CLI commands.
func SortProjects(projects []core.Project) {
	sort.Slice(projects, func(i, j int) bool {
		leftName := strings.ToLower(strings.TrimSpace(projects[i].Name))
		rightName := strings.ToLower(strings.TrimSpace(projects[j].Name))
		if leftName == "" {
			leftName = strings.ToLower(projects[i].ProjectID)
		}
		if rightName == "" {
			rightName = strings.ToLower(projects[j].ProjectID)
		}
		if leftName == rightName {
			return strings.ToLower(projects[i].ProjectID) < strings.ToLower(projects[j].ProjectID)
		}
		return leftName < rightName
	})
}

// StdinAvailable reports whether the given reader is a non-terminal file.
func StdinAvailable(in io.Reader) bool {
	file, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}
