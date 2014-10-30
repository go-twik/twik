package twik_test

import (
	. "gopkg.in/check.v1"
	"gopkg.in/twik.v1"
)

func (S) BenchmarkParse0(c *C) {
	fset := twik.NewFileSet()
	var err error
	for i := 0; i < c.N; i++ {
		_, err = twik.ParseString(fset, "", "0")
	}
	c.StopTimer()
	c.Assert(err, IsNil)
}

func (S) BenchmarkEval0(c *C) {
	fset := twik.NewFileSet()
	node, err := twik.ParseString(fset, "", "0")
	c.Assert(err, IsNil)
	var value interface{}
	c.ResetTimer()
	for i := 0; i < c.N; i++ {
		value, err = twik.NewScope(fset).Eval(node)
	}
	c.StopTimer()
	c.Assert(err, IsNil)
	c.Assert(value, Equals, int64(0))
}


func (S) BenchmarkParseFib(c *C) {
	fset := twik.NewFileSet()
	var err error
	c.ResetTimer()
	for i := 0; i < c.N; i++ {
		_, err = twik.ParseString(fset, "", "(func fib (n) (if (== n 0) 0 (if (== n 1) 1 (+ (fib (- n 1)) (fib (- n 2))))))")
	}
	c.StopTimer()
	c.Assert(err, IsNil)
}

func (S) BenchmarkEvalFib10(c *C) {
	fset := twik.NewFileSet()
	node, err := twik.ParseString(fset, "", "(func fib (n) (if (== n 0) 0 (if (== n 1) 1 (+ (fib (- n 1)) (fib (- n 2)))))) (fib 10)")
	c.Assert(err, IsNil)
	var value interface{}
	c.ResetTimer()
	for i := 0; i < c.N; i++ {
		value, err = twik.NewScope(fset).Eval(node)
	}
	c.StopTimer()
	c.Assert(err, IsNil)
	c.Assert(value, NotNil)
}
