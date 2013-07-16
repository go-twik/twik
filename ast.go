package twik

import (
	"launchpad.net/twik/ast"
)

// NewFileSet returns a new FileSet to hold positioning information
// for a set of parsed twik sources.
func NewFileSet() *ast.FileSet {
	return ast.NewFileSet()
}

// Parse parses a byte slice containing twik code and returns
// the resulting parsed tree.
//
// Positioning information for the parsed code will be stored in
// fset under the given name.
func Parse(fset *ast.FileSet, name string, code []byte) (ast.Node, error) {
	return ast.ParseString(fset, name, string(code))
}

// ParseString parses a string containing twik code and returns
// the resulting parsed tree.
//
// Positioning information for the parsed code will be stored in
// fset under the given name.
func ParseString(fset *ast.FileSet, name string, code string) (ast.Node, error) {
	return ast.ParseString(fset, name, string(code))
}

