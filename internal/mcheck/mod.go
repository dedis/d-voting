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

// iterateOverLines iterates over the lines of a comment and
// checks if the line is too long or contains a prefix
// that should be ignored
func iterateOverLines(pass *analysis.Pass, lines []string, c *ast.Comment) {
	for j := 0; j < len(lines); j++ {
		line := lines[j]
		if checkPrefixes(line) {
			continue
		}
		ifTooLong(line, pass, c)
		if strings.HasPrefix(line, NoLint) {
			// Skip next comment for block comment.
			j++
		}
	}
}

// checkPrefixes checks if the line starts with any of the following
// prefixes:
//
//	`go:generate`
//	`http://`
//	`https://`
//
// If it does, it returns `true`. Otherwise, it returns `false`
func checkPrefixes(line string) bool {
	return strings.HasPrefix(line, "//go:generate") ||
		strings.HasPrefix(line, "// http://") ||
		strings.HasPrefix(line, "// https://")
}

// ifTooLong reports a comment if it's too long
func ifTooLong(line string, pass *analysis.Pass, c *ast.Comment) {
	if len(line) > MaxLen {
		pass.Reportf( // `c` is a comment.
			c.Pos(), "Comment too long: %s (%d)",
			line, len(line))
	}
}

// runComment loops over all the files in the package, and for each file it
// loops over all the comments in the file, and for each comment it loops
// over all the lines in the comment, and for each line it checks
// if the line is too long
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
				iterateOverLines(pass, lines, c)
				isFirst = false
				if strings.HasPrefix(c.Text, NoLint) {
					// Skip next comment for block comment.
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
