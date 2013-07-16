package twik_test

import (
	"github.com/kr/pretty"
	. "launchpad.net/gocheck"
	"launchpad.net/twik"
)

func (S) TestParser(c *C) {
	for _, test := range parserTests {
		fset := twik.NewFileSet()
		root, err := twik.ParseString(fset, "", test.code)
		if e, ok := test.value.(error); ok {
			c.Assert(err, ErrorMatches, e.Error())
			c.Assert(root, IsNil)
		} else {
			c.Assert(err, IsNil)
			if !c.Check(root.(*twik.Root).Nodes, DeepEquals, test.value) {
				c.Logf("Obtained: %# v", pretty.Formatter(root.(*twik.Root).Nodes))
				c.Logf("Expected: %# v", pretty.Formatter(test.value))
				c.FailNow()
			}
		}
	}
}

var parserTests = []struct {
	code  string
	value interface{}
}{
	{
		`1`,
		[]twik.Node{
			&twik.Int{Input: "1", InputPos: 1, Value: 1},
		},
	}, {
		`-1`,
		[]twik.Node{
			&twik.Int{Input: "-1", InputPos: 1, Value: -1},
		},
	}, {
		` 1 `,
		[]twik.Node{
			&twik.Int{Input: "1", InputPos: 2, Value: 1},
		},
	}, {
		`0x10`,
		[]twik.Node{
			&twik.Int{Input: "0x10", InputPos: 1, Value: 16},
		},
	}, {
		`0n10`,
		errorf(".*: invalid int literal: 0n10"),
	}, {
		`'a'`,
		[]twik.Node{
			&twik.Int{Input: "'a'", InputPos: 1, Value: 'a'},
		},
	}, {
		`'\''`,
		[]twik.Node{
			&twik.Int{Input: `'\''`, InputPos: 1, Value: '\''},
		},
	}, {
		`'`,
		errorf(".*: invalid single quote"),
	}, {
		`''`,
		errorf(".*: invalid single quote"),
	}, {
		` 1.0 `,
		[]twik.Node{
			&twik.Float{Input: "1.0", InputPos: 2, Value: 1},
		},
	}, {
		`()`,
		[]twik.Node{
			&twik.List{LParens: 1, RParens: 2},
		},
	}, {
		` ( ) `,
		[]twik.Node{
			&twik.List{LParens: 2, RParens: 4},
		},
	}, {
		`"foo\"bar"`,
		[]twik.Node{
			&twik.String{Input: `"foo\"bar"`, InputPos: 1, Value: "foo\"bar"},
		},
	}, {
		` "foo" `,
		[]twik.Node{
			&twik.String{Input: `"foo"`, InputPos: 2, Value: "foo"},
		},
	}, {
		` "foo `,
		errorf(`.*: unclosed string literal: "foo `),
	}, {
		`"\m"`,
		errorf(`.*: invalid string literal: "\\m"`),
	}, {
		`(+ 1 (- 2 3) 4)`,
		[]twik.Node{
			&twik.List{
				LParens: 1,
				Nodes: []twik.Node{
					&twik.Symbol{Name: "+", NamePos: 2},
					&twik.Int{Input: "1", InputPos: 4, Value: 1},
					&twik.List{
						LParens: 6,
						Nodes: []twik.Node{
							&twik.Symbol{Name: "-", NamePos: 7},
							&twik.Int{Input: "2", InputPos: 9, Value: 2},
							&twik.Int{Input: "3", InputPos: 11, Value: 3},
						},
						RParens: 12,
					},
					&twik.Int{Input: "4", InputPos: 14, Value: 4},
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
		[]twik.Node{
			&twik.Int{Input: "1", InputPos: 11, Value: 1},
		},
	}, {
		"(; Comment\n1)",
		[]twik.Node{
			&twik.List{
				LParens: 1,
				Nodes: []twik.Node{
					&twik.Int{Input: "1", InputPos: 12, Value: 1},
				},
				RParens: 13,
			},
		},
	},
}
