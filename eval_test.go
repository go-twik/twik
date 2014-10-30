package twik_test

import (
	"fmt"
	"testing"

	. "gopkg.in/check.v1"
	"gopkg.in/twik.v1"
)

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(S{})

type S struct{}

func (S) TestEval(c *C) {
	for _, test := range evalList {
		fset := twik.NewFileSet()
		node, err := twik.ParseString(fset, "", test.code)
		c.Assert(err, IsNil)
		scope := twik.NewScope(fset)
		scope.Create("sprintf", sprintfFn)
		scope.Create("list", listFn)
		scope.Create("append", appendFn)
		value, err := scope.Eval(node)
		if e, ok := test.value.(error); ok {
			c.Assert(err, ErrorMatches, e.Error(), Commentf("Code: %s", test.code))
			c.Assert(value, IsNil)
		} else {
			tvalue := test.value
			if i, ok := tvalue.(int); ok {
				tvalue = int64(i)
			}
			c.Assert(err, IsNil, Commentf("Code: %s", test.code))
			c.Assert(value, DeepEquals, tvalue, Commentf("Code: %s", test.code))
		}
	}
}

func sprintfFn(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("sprintf takes at least one format argument")
	}
	format, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("sprintf takes format string as first argument")
	}
	return fmt.Sprintf(format, args[1:]...), nil
}

func listFn(args []interface{}) (interface{}, error) {
	return args, nil
}

func appendFn(args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("append takes list as first argument")
	}
	return append(list, args[1:]...), nil
}

func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

var evalList = []struct {
	code  string
	value interface{}
}{
	// some basics
	{
		`1`,
		1,
	}, {
		`1.0`,
		1.0,
	}, {
		`0x10`,
		16,
	}, {
		`010`,
		8,
	}, {
		`"foo\"bar"`,
		`foo"bar`,
	}, {
		`foo`,
		errorf("twik source:1:1: undefined symbol: foo"),
	}, {
		`(1)`,
		errorf(`twik source:1:2: cannot use 1 as a function`),
	}, {
		`true`,
		true,
	}, {
		`false`,
		false,
	}, {
		`nil`,
		nil,
	}, {
		`1 2 3`,
		3,
	},

	// error
	{
		"(\nerror \"error message\")",
		errorf("twik source:2:1: error message"),
	}, {
		`(error)`,
		errorf("twik source:1:2: error function takes a single string argument"),
	}, {
		`(error 1)`,
		errorf("twik source:1:2: error function takes a single string argument"),
	}, {
		`(error "foo" 2)`,
		errorf("twik source:1:2: error function takes a single string argument"),
	},

	// +
	{
		`(+)`,
		0,
	}, {
		`(+ 1)`,
		1,
	}, {
		`(+ 1 2)`,
		3,
	}, {
		`(+ 1 (+ 2 3))`,
		6,
	}, {
		`(+ "123")`,
		errorf(`twik source:1:2: cannot sum "123"`),
	}, {
		`(+ 1.5)`,
		1.5,
	}, {
		`(+ 1.5 1.5)`,
		3.0,
	}, {
		`(+ 1.5 1)`,
		2.5,
	}, {
		`(+ 1 1.5)`,
		2.5,
	},

	// -
	{
		`(-)`,
		errorf(`twik source:1:2: function "-" takes one or more arguments`),
	}, {
		`(- 1)`,
		-1,
	}, {
		`(- 10 1)`,
		9,
	}, {
		`(- 10 1 2)`,
		7,
	}, {
		`(- 10 (- 2 1))`,
		9,
	}, {
		`(- "123")`,
		errorf(`twik source:1:2: cannot subtract "123"`),
	}, {
		`(- 1.5)`,
		-1.5,
	}, {
		`(- 2.0 1.5)`,
		0.5,
	}, {
		`(- 1.5 1)`,
		0.5,
	}, {
		`(- 1 1.5)`,
		-0.5,
	},

	// *
	{
		`(*)`,
		1,
	}, {
		`(* 1)`,
		1,
	}, {
		`(* 2 3 4)`,
		24,
	}, {
		`(* 2 (* 3 4))`,
		24,
	}, {
		`(* "123")`,
		errorf(`twik source:1:2: cannot multiply "123"`),
	}, {
		`(* 1.5)`,
		1.5,
	}, {
		`(* 2.0 1.5)`,
		3.0,
	}, {
		`(* 1.5 1)`,
		1.5,
	}, {
		`(* 1 1.5)`,
		1.5,
	},

	// /
	{
		`(/)`,
		errorf(`twik source:1:2: function "/" takes two or more arguments`),
	}, {
		`(/ 1)`,
		errorf(`twik source:1:2: function "/" takes two or more arguments`),
	}, {
		`(/ 10 2)`,
		5,
	}, {
		`(/ 30 3 2)`,
		5,
	}, {
		`(/ 30 (/ 10 2))`,
		6,
	}, {
		`(/ 10 "123")`,
		errorf(`twik source:1:2: cannot divide with "123"`),
	}, {
		`(/ 10.0 2.0)`,
		5.0,
	}, {
		`(/ 10.0 2)`,
		5.0,
	}, {
		`(/ 10 2.0)`,
		5.0,
	},


	// ==
	{
		`(== "a" "a")`,
		true,
	}, {
		`(== "a" "b")`,
		false,
	}, {
		`(== 42 42)`,
		true,
	}, {
		`(== 42 43)`,
		false,
	}, {
		`(== 42 "a")`,
		false,
	}, {
		`(== 42 42.0)`,
		false,
	}, {
		`(== 1 2 3)`,
		errorf("twik source:1:2: == takes two values"),
	}, {
		`(==)`,
		errorf("twik source:1:2: == takes two values"),
	},

	// !=
	{
		`(!= "a" "a")`,
		false,
	}, {
		`(!= "a" "b")`,
		true,
	}, {
		`(!= 42 42)`,
		false,
	}, {
		`(!= 42 43)`,
		true,
	}, {
		`(!= 42 "a")`,
		true,
	}, {
		`(!= 42 42.0)`,
		true,
	}, {
		`(!= 1 2 3)`,
		errorf("twik source:1:2: != takes two values"),
	}, {
		`(!=)`,
		errorf("twik source:1:2: != takes two values"),
	},


	// or
	{
		`(or)`,
		false,
	}, {
		`(or false 1 2 (error "must not get here"))`,
		1,
	}, {
		`(or (error "boom") 1 2 3)`,
		errorf("twik source:1:6: boom"),
	},

	// and
	{
		`(and)`,
		true,
	}, {
		`(and 1 2 3)`,
		3,
	}, {
		`(and false (error "must not get here"))`,
		false,
	}, {
		`(and (error "boom") true)`,
		errorf("twik source:1:7: boom"),
	},

	// var
	{
		`(var x (+ 1 2)) x`,
		3,
	}, {
		`(var x) x`,
		nil,
	}, {
		`(var x 1 2)`,
		errorf("twik source:1:2: var takes one or two arguments"),
	}, {
		`(var)`,
		errorf("twik source:1:2: var takes one or two arguments"),
	}, {
		"(var x)\n(var x)",
		errorf("twik source:2:2: symbol already defined in current scope: x"),
	},

	// set
	{
		`(var x) (set x 2) (+ x 3)`,
		5,
	}, {
		`(set x 1)`,
		errorf("twik source:1:2: cannot set undefined symbol: x"),
	}, {
		`(var x) (set x 1 2)`,
		errorf(`twik source:1:10: function "set" takes two arguments`),
	}, {
		`(var x) (set x)`,
		errorf(`twik source:1:10: function "set" takes two arguments`),
	}, {
		`(var x) (set)`,
		errorf(`twik source:1:10: function "set" takes two arguments`),
	},

	// do
	{
		`(do)`,
		nil,
	}, {
		`(do 1 2 3)`,
		3,
	}, {
		`(var x 1) (do (set x 2) x)`,
		2,
	}, {
		`(var x 1) (do (set x 2)) x`,
		2,
	}, {
		`(var x 1) (do (var x) (set x 2) x)`,
		2,
	}, {
		`(var x 1) (do (var x) (set x 2)) x`,
		1,
	},

	// func
	{
		`((func (a b) (+ a b)) 1 2)`,
		3,
	}, {
		`(var add (do (var x 0) (func (n) (set x (+ x n)) x))) (add 1) (add 2)`,
		3,
	}, {
		`(func add (a b) (+ a b)) (add 1 2)`,
		3,
	}, {
		`(func)`,
		errorf("twik source:1:2: func takes three or more arguments"),
	}, {
		`(func x)`,
		errorf("twik source:1:2: func takes three or more arguments"),
	}, {
		`(func 1 2)`,
		errorf("twik source:1:2: func takes a list of parameters"),
	}, {
		`(func f 2)`,
		errorf("twik source:1:2: func takes a list of parameters"),
	}, {
		`(func f (a)) (f 1 2)`,
		errorf(`twik source:1:2: func takes a body sequence`),
	}, {
		"(var f (func (a) 1))\n(f 1 2)",
		errorf(`twik source:2:2: anonymous function takes one argument`),
	}, {
		"(func f () 1)\n(f 1)",
		errorf(`twik source:2:2: function "f" takes no arguments`),
	}, {
		"(func f (a) 1)\n(f 1 2)",
		errorf(`twik source:2:2: function "f" takes one argument`),
	}, {
		"(func f (a b) 1)\n(f 1)",
		errorf(`twik source:2:2: function "f" takes 2 arguments`),
	},

	// if
	{
		`(if true 1)`,
		1,
	}, {
		`(if 0 1)`,
		1,
	}, {
		`(if false 1)`,
		false,
	}, {
		`(if false 1 2)`,
		2,
	}, {
		`(if)`,
		errorf(`twik source:1:2: function "if" takes two or three arguments`),
	}, {
		`(if 1)`,
		errorf(`twik source:1:2: function "if" takes two or three arguments`),
	},

	// for
	{
		`(for 1 2 3)`,
		errorf("twik source:1:2: for takes four or more arguments"),
	}, {
		`(for (error "init") (error "test") (error "step") (error "code"))`,
		errorf("twik source:1:7: init"),
	}, {
		`(for () (error "test") (error "step") (error "code"))`,
		errorf("twik source:1:10: test"),
	}, {
		`(for () () (error "step") (error "code"))`,
		errorf("twik source:1:28: code"),
	}, {
		`(for () () (error "step") ())`,
		errorf("twik source:1:13: step"),
	}, {
		`(for (var i 0) false () ()) i`,
		errorf("twik source:1:29: undefined symbol: i"),
	}, {
		`(var x 0) (for (var i 0) (!= i 4) (set i (+ i 1)) (set x (+ x i)) (* 2 x))`,
		12,
	},

	// range
	{
		`(range 1 2)`,
		errorf("twik source:1:2: range takes three or more arguments"),
	}, {
		`(range 1 2 3)`,
		errorf(`twik source:1:2: range takes var name or \(i elem\) var name pair as first argument`),
	}, {
		`(range i 0 ()) i`,
		errorf("twik source:1:16: undefined symbol: i"),
	}, {
		`(var x 0) (range i 4 (set x (+ x i)) (* 2 x))`,
		12,
	}, {
		`(var l ()) (range (i e) (list "A" "B" "C") (set l (append l i e))) l`,
		[]interface {}{0, "A", 1, "B", 2, "C"},
	},


	// calling of custom functions
	{
		`(sprintf "Value: %.02f" 1.0)`,
		"Value: 1.00",
	},
}
