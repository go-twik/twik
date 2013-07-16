package ast_test

import (
	"fmt"
	"github.com/kr/pretty"
	. "launchpad.net/gocheck"
	"launchpad.net/twik/ast"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(S{})

type S struct{}

func (S) TestParser(c *C) {
	for _, test := range parserTests {
		fset := ast.NewFileSet()
		root, err := ast.ParseString(fset, "", test.code)
		if e, ok := test.value.(error); ok {
			c.Assert(err, ErrorMatches, e.Error())
			c.Assert(root, IsNil)
		} else {
			c.Assert(err, IsNil)
			if !c.Check(root.(*ast.Root).Nodes, DeepEquals, test.value) {
				c.Logf("Obtained: %# v", pretty.Formatter(root.(*ast.Root).Nodes))
				c.Logf("Expected: %# v", pretty.Formatter(test.value))
				c.FailNow()
			}
		}
	}
}

func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

var parserTests = []struct {
	code  string
	value interface{}
}{
	{
		`1`,
		[]ast.Node{
			&ast.Int{Input: "1", InputPos: 1, Value: 1},
		},
	}, {
		`-1`,
		[]ast.Node{
			&ast.Int{Input: "-1", InputPos: 1, Value: -1},
		},
	}, {
		` 1 `,
		[]ast.Node{
			&ast.Int{Input: "1", InputPos: 2, Value: 1},
		},
	}, {
		`0x10`,
		[]ast.Node{
			&ast.Int{Input: "0x10", InputPos: 1, Value: 16},
		},
	}, {
		`0n10`,
		errorf(".*: invalid int literal: 0n10"),
	}, {
		`'a'`,
		[]ast.Node{
			&ast.Int{Input: "'a'", InputPos: 1, Value: 'a'},
		},
	}, {
		`'\''`,
		[]ast.Node{
			&ast.Int{Input: `'\''`, InputPos: 1, Value: '\''},
		},
	}, {
		`'`,
		errorf(".*: invalid single quote"),
	}, {
		`''`,
		errorf(".*: invalid single quote"),
	}, {
		` 1.0 `,
		[]ast.Node{
			&ast.Float{Input: "1.0", InputPos: 2, Value: 1},
		},
	}, {
		`()`,
		[]ast.Node{
			&ast.List{LParens: 1, RParens: 2},
		},
	}, {
		` ( ) `,
		[]ast.Node{
			&ast.List{LParens: 2, RParens: 4},
		},
	}, {
		`"foo\"bar"`,
		[]ast.Node{
			&ast.String{Input: `"foo\"bar"`, InputPos: 1, Value: "foo\"bar"},
		},
	}, {
		` "foo" `,
		[]ast.Node{
			&ast.String{Input: `"foo"`, InputPos: 2, Value: "foo"},
		},
	}, {
		` "foo `,
		errorf(`.*: unclosed string literal: "foo `),
	}, {
		`"\m"`,
		errorf(`.*: invalid string literal: "\\m"`),
	}, {
		`(+ 1 (- 2 3) 4)`,
		[]ast.Node{
			&ast.List{
				LParens: 1,
				Nodes: []ast.Node{
					&ast.Symbol{Name: "+", NamePos: 2},
					&ast.Int{Input: "1", InputPos: 4, Value: 1},
					&ast.List{
						LParens: 6,
						Nodes: []ast.Node{
							&ast.Symbol{Name: "-", NamePos: 7},
							&ast.Int{Input: "2", InputPos: 9, Value: 2},
							&ast.Int{Input: "3", InputPos: 11, Value: 3},
						},
						RParens: 12,
					},
					&ast.Int{Input: "4", InputPos: 14, Value: 4},
				},
				RParens: 15,
			},
		},
	},

	{
		"(a\nb\nc",
		errorf(`twik source:3:2: missing \)`),
	}, {
		"(a\nb\n 1n \n)",
		errorf(`twik source:3:2: invalid int literal: 1n`),
	}, {
		"1n",
		errorf(`twik source:1:1: invalid int literal: 1n`),
	}, {
		"; Comment\n1",
		[]ast.Node{
			&ast.Int{Input: "1", InputPos: 11, Value: 1},
		},
	}, {
		"(; Comment\n1)",
		[]ast.Node{
			&ast.List{
				LParens: 1,
				Nodes: []ast.Node{
					&ast.Int{Input: "1", InputPos: 12, Value: 1},
				},
				RParens: 13,
			},
		},
	},
}
