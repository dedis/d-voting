package main

// This package provides a custom checks for "go vet".
// It can be used like the following:
//  `go build && go vet -vettool=./check -commentLen -ifInit ./...`

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/unitchecker"
)

// NoLint is a command to disable linting for the next line.
const NoLint = "// @nolint-next-line"

// MaxLen is the maximum length of a comment
var MaxLen = 80

// This check verifies that no comments exceed the "MaxLen" length. It ignores
// files that have as first comment a "// Code genereated..." comment and it
// ignores comments that start with "//go:generate". It also ignores lines that
// starts with a link, ie start with "http(s)://". One can also ignore a next
// comment with "// @nolint-next-line".
var commentLenAnalyzer = &analysis.Analyzer{
	Name: "commentLen",
	Doc:  "checks the lengths of comments",
	Run:  runComment,
}

// This check ensures that no if has an initialization statement
var ifInitAnalyzer = &analysis.Analyzer{
	Name: "ifInit",
	Doc:  "checks that no if with an initialization statement are used",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: runIfInitCheck,
}

func main() {
	unitchecker.Main(
		commentLenAnalyzer,
		ifInitAnalyzer,
	)
}

// run parses all the comments in ast.File
func runComment(pass *analysis.Pass) (interface{}, error) {
fileLoop:
	for _, file := range pass.Files {
		isFirst := true
		for _, cg := range file.Comments {
			for i := 0; i < len(cg.List); i++ {
				c := cg.List[i]

				if isFirst && strings.HasPrefix(c.Text, "// Code generated") {
					continue fileLoop
				}
				// in case of /* */ comment there might be multiple lines
				lines := strings.Split(c.Text, "\n")
				for j := 0; j < len(lines); j++ {
					line := lines[j]

					if strings.HasPrefix(line, "//go:generate") {
						continue
					}
					if strings.HasPrefix(line, "// http://") || strings.HasPrefix(line, "// https://") {
						continue
					}
					if len(line) > MaxLen {
						pass.Reportf(c.Pos(), "Comment too long: %s (%d)",
							line, len(line))
					}
					if strings.HasPrefix(line, NoLint) {
						// Skip next comment for block comment.
						j++
					}
				}

				isFirst = false

				if strings.HasPrefix(c.Text, NoLint) {
					// Skip next comment for one-line comment.
					i++
				}
			}
		}
	}

	return nil, nil
}

// runIfInitCheck parses all the if statement and checks if there is an
// initialization statement used.
func runIfInitCheck(pass *analysis.Pass) (interface{}, error) {
fileLoop:
	for _, file := range pass.Files {
		// We ignore generated files
		if len(file.Comments) != 0 {
			cg := file.Comments[0]
			if len(cg.List) != 0 {
				comment := cg.List[0]
				if strings.HasPrefix(comment.Text, "// Code generated") {
					continue fileLoop
				}
			}
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.IfStmt:
				if x.Init != nil {
					pass.Reportf(x.Pos(), "Please do not do initialization "+
						"in if statement")
				}
			}
			return true
		})
	}

	return nil, nil
}
