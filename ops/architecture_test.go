package ops

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFrontendDependencyBoundaries(t *testing.T) {
	for _, directory := range []string{".", "../mcp"} {
		err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			for _, im := range file.Imports {
				name, err := strconv.Unquote(im.Path.Value)
				if err != nil {
					return err
				}
				if name == "github.com/yumauri/fbrcm/cli" || strings.HasPrefix(name, "github.com/yumauri/fbrcm/cli/") {
					t.Errorf("%s imports CLI frontend %s", path, name)
				}
				if (strings.HasPrefix(path, "workflows/") || strings.HasPrefix(path, "shared/") || strings.HasPrefix(path, "invocation/") || strings.HasPrefix(path, "../mcp/server/")) && name == "github.com/spf13/cobra" {
					t.Errorf("%s depends on Cobra execution state", path)
				}
			}
			if filepath.Base(path) != "init.go" {
				ast.Inspect(file, func(node ast.Node) bool {
					if selector, ok := node.(*ast.SelectorExpr); ok && (selector.Sel.Name == "ExecuteC" || selector.Sel.Name == "SetArgs" || selector.Sel.Name == "ParseFlags") {
						t.Errorf("%s reintroduces command execution through %s", path, selector.Sel.Name)
					}
					return true
				})
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
