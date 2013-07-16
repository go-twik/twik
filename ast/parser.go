package ast

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pos is a position marker within a file set. Use the FileSet's PosInfo
// method to obtain human-oriented details for the position.
type Pos int

// The Node interface is implemented by all AST nodes that result
// from parsing twik code.
type Node interface {
	Pos() Pos
	End() Pos
}

// Int represents an integer literal in parsed twik code.
type Int struct {
	Input    string
	InputPos Pos
	Value    int64
}

func (l *Int) Pos() Pos { return l.InputPos }
func (l *Int) End() Pos { return l.InputPos + Pos(len(l.Input)) }

// Float represents a float literal in parsed twik code.
type Float struct {
	Input    string
	InputPos Pos
	Value    float64
}

func (l *Float) Pos() Pos { return l.InputPos }
func (l *Float) End() Pos { return l.InputPos + Pos(len(l.Input)) }

// String represents a string literal in parsed twik code.
type String struct {
	Input    string
	InputPos Pos
	Value    string
}

func (l *String) Pos() Pos { return l.InputPos }
func (l *String) End() Pos { return l.InputPos + Pos(len(l.Input)) }

// Symbol represents a symbol in parsed twik code.
type Symbol struct {
	Name    string
	NamePos Pos
}

func (s *Symbol) Pos() Pos { return s.NamePos }
func (s *Symbol) End() Pos { return s.NamePos + Pos(len(s.Name)) }

// List represents a list of entries from parsed twik code.
type List struct {
	LParens Pos
	RParens Pos
	Nodes   []Node
}

func (s *List) Pos() Pos { return s.LParens }
func (s *List) End() Pos { return s.RParens + 1 }

// Root represents the root of parsed twik code.
type Root struct {
	First Pos
	After Pos
	Nodes []Node
}

func (s *Root) Pos() Pos { return s.First }
func (s *Root) End() Pos { return s.After }

// Parse parses a byte slice containing twik code and returns
// the resulting parsed tree.
//
// Positioning information for the parsed code will be stored in
// fset under the given name.
func Parse(fset *FileSet, name string, code []byte) (Node, error) {
	return ParseString(fset, name, string(code))
}

// ParseString parses a string containing twik code and returns
// the resulting parsed tree.
//
// Positioning information for the parsed code will be stored in
// fset under the given name.
func ParseString(fset *FileSet, name string, code string) (Node, error) {
	base := fset.nextBase()
	fset.files = append(fset.files, file{name: name, code: code, base: base})

	p := parser{fset: fset, code: code, base: base}
	root := Root{First: p.pos(0)}
	node, err := p.next()
	for err == nil {
		root.Nodes = append(root.Nodes, node)
		node, err = p.next()
	}
	if err != io.EOF {
		if err == errOpened || err == errClosed {
			return nil, p.ierrorf(p.i, "%v", err)
		}
		return nil, err
	}
	root.After = p.pos(p.i)
	return &root, nil
}

type parser struct {
	fset *FileSet
	code string
	base Pos
	i    int
}

var errClosed = errors.New("unexpected )")
var errOpened = errors.New("missing )")

type closedError struct {
}

func (p *parser) pos(i int) Pos {
	return p.base + Pos(i)
}

func (p *parser) ierrorf(i int, format string, args ...interface{}) error {
	pinfo := p.fset.PosInfo(p.pos(i))
	return fmt.Errorf("%s %s", pinfo, fmt.Sprintf(format, args...))
}

func (p *parser) next() (Node, error) {
	if p.i == len(p.code) {
		return nil, io.EOF
	}

	r, size := utf8.DecodeRuneInString(p.code[p.i:])
	for r == ';' || unicode.IsSpace(r) {
		p.i += size
		if r == ';' {
			for p.i < len(p.code) && r != '\n' {
				r, size = utf8.DecodeRuneInString(p.code[p.i:])
				p.i += size
			}
		}
		if p.i == len(p.code) {
			return nil, io.EOF
		}
		r, size = utf8.DecodeRuneInString(p.code[p.i:])
	}
	start := p.i
	p.i += size

	if r == ')' {
		return nil, errClosed
	}
	if r == '(' {
		var nodes []Node
		for {
			node, err := p.next()
			if err == errClosed {
				break
			}
			if err == io.EOF {
				return nil, errOpened
			}
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
		list := &List{
			LParens: p.pos(start),
			RParens: p.pos(p.i - 1),
			Nodes:   nodes,
		}
		return list, nil
	}

	if r == '-' && p.i < len(p.code) {
		r, size = utf8.DecodeRuneInString(p.code[p.i:])
		if r >= '0' && r <= '9' {
			// It's a digit; consume minus now and fall onto number case.
			p.i += size
		}
	}

	// int, float
	if r >= '0' && r <= '9' {
		dot := false
		for p.i < len(p.code) {
			r, size = utf8.DecodeRuneInString(p.code[p.i:])
			if r == '.' {
				dot = true
			} else if r == ')' || unicode.IsSpace(r) {
				break
			}
			p.i += size
		}
		input := p.code[start:p.i]
		if dot {
			value, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return nil, p.ierrorf(start, "invalid float literal: %s", input)
			}
			return &Float{Input: input, InputPos: p.pos(start), Value: value}, nil
		} else {
			value, err := strconv.ParseInt(input, 0, 64)
			if err != nil {
				return nil, p.ierrorf(start, "invalid int literal: %s", input)
			}
			return &Int{Input: input, InputPos: p.pos(start), Value: value}, nil
		}
	}

	if r == '\'' {
		var c rune
		if p.i < len(p.code) {
			c, size = utf8.DecodeRuneInString(p.code[p.i:])
			p.i += size
			if c == '\\' && p.i < len(p.code) {
				c, size = utf8.DecodeRuneInString(p.code[p.i:])
				p.i += size
			} else if c == '\'' {
				return nil, p.ierrorf(start, "invalid single quote")
			}
		}
		if p.i == len(p.code) {
			return nil, p.ierrorf(start, "invalid single quote")
		}
		r, size = utf8.DecodeRuneInString(p.code[p.i:])
		p.i += size
		if r != '\'' {
			return nil, p.ierrorf(start, "unclosed single quote")
		}
		return &Int{Input: p.code[start:p.i], InputPos: p.pos(start), Value: int64(c)}, nil
	}

	// string
	if r == '"' {
		escaped := false
		for {
			if p.i == len(p.code) {
				return nil, p.ierrorf(start, "unclosed string literal: %s", p.code[start:])
			}
			r, size = utf8.DecodeRuneInString(p.code[p.i:])
			p.i += size
			if r == '"' && !escaped {
				break
			}
			escaped = r == '\\'
		}
		input := p.code[start:p.i]
		value, err := strconv.Unquote(input)
		if err != nil {
			return nil, p.ierrorf(start, "invalid string literal: %s", input)
		}
		return &String{Input: input, InputPos: p.pos(start), Value: value}, nil
	}

	// symbol
	for p.i < len(p.code) {
		r, size = utf8.DecodeRuneInString(p.code[p.i:])
		if r == ')' || unicode.IsSpace(r) {
			break
		}
		p.i += size
	}
	symbol := &Symbol{
		Name:    p.code[start:p.i],
		NamePos: p.pos(start),
	}
	return symbol, nil
}

// NewFileSet returns a new FileSet.
func NewFileSet() *FileSet {
	return &FileSet{}
}

// FileSet holds positioning information for parsed twik code. 
type FileSet struct {
	files []file
}

type file struct {
	name string
	code string
	base Pos
}

func (fset *FileSet) nextBase() Pos {
	if len(fset.files) == 0 {
		return 1
	} else {
		last := fset.files[len(fset.files)-1]
		return last.base + Pos(len(last.code)) + 1
	}
}

// PosInfo returns the line and column for pos, and the name the
// file containing that position was parsed with.
func (fset *FileSet) PosInfo(pos Pos) *PosInfo {
	pinfo := &PosInfo{}
	for _, f := range fset.files {
		if pos <= f.base+Pos(len(f.code)) {
			offset := int(pos - f.base)
			code := f.code[:offset]
			pinfo.Name = f.name
			pinfo.Line = 1 + strings.Count(code, "\n")
			if i := strings.LastIndex(code, "\n"); i >= 0 {
				pinfo.Column = offset - i
			} else {
				pinfo.Column = 1 + len(code)
			}
		}
	}
	return pinfo
}

// PosInfo holds human-oriented positioning details about a Pos.
type PosInfo struct {
	Name   string
	Line   int
	Column int
}

func (info *PosInfo) String() string {
	name := "twik source"
	if info.Name != "" {
		name = info.Name
	}
	return fmt.Sprintf("%s:%d:%d:", name, info.Line, info.Column)
}
