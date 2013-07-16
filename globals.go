package twik

import (
	"errors"
	"fmt"
)

var defaultGlobals = []struct {
	name  string
	value interface{}
}{
	{"true", true},
	{"false", false},
	{"nil", nil},
	{"error", errorFn},
	{"==", eqFn},
	{"!=", neFn},
	{"+", plusFn},
	{"-", minusFn},
	{"*", mulFn},
	{"/", divFn},
	{"or", orFn},
	{"and", andFn},
	{"if", ifFn},
	{"var", varFn},
	{"set", setFn},
	{"do", doFn},
	{"func", funcFn},
	//{"for", forFn},
}

func errorFn(args []interface{}) (value interface{}, err error) {
	if len(args) == 1 {
		if s, ok := args[0].(string); ok {
			return nil, errors.New(s)
		}
	}
	return nil, errors.New("error function takes a single string argument")
}

func eqFn(args []interface{}) (value interface{}, err error) {
	if len(args) != 2 {
		return nil, errors.New("== expects two values")
	}
	return args[0] == args[1], nil
}

func neFn(args []interface{}) (value interface{}, err error) {
	if len(args) != 2 {
		return nil, errors.New("!= expects two values")
	}
	return args[0] != args[1], nil
}

func plusFn(args []interface{}) (value interface{}, err error) {
	var resi int64
	var resf float64
	var f bool
	for _, arg := range args {
		switch arg := arg.(type) {
		case int64:
			resi += arg
			resf += float64(arg)
		case float64:
			resf += arg
			f = true
		default:
			return nil, fmt.Errorf("cannot sum %#v", arg)
		}
	}
	if f {
		return resf, nil
	}
	return resi, nil
}

func minusFn(args []interface{}) (value interface{}, err error) {
	if len(args) == 0 {
		return nil, fmt.Errorf(`function "-" takes one or more arguments`)
	}
	var resi int64
	var resf float64
	var f bool
	for i, arg := range args {
		switch arg := arg.(type) {
		case int64:
			if i == 0 && len(args) > 1 {
				resi = arg
				resf = float64(arg)
			} else {
				resi -= arg
				resf -= float64(arg)
			}
		case float64:
			if i == 0 && len(args) > 1 {
				resf = arg
			} else {
				resf -= arg
			}
			f = true
		default:
			return nil, fmt.Errorf("cannot subtract %#v", arg)
		}
	}
	if f {
		return resf, nil
	}
	return resi, nil
}

func mulFn(args []interface{}) (value interface{}, err error) {
	var resi = int64(1)
	var resf = float64(1)
	var f bool
	for _, arg := range args {
		switch arg := arg.(type) {
		case int64:
			resi *= arg
			resf *= float64(arg)
		case float64:
			resf *= arg
			f = true
		default:
			return nil, fmt.Errorf("cannot multiply %#v", arg)
		}
	}
	if f {
		return resf, nil
	}
	return resi, nil
}

func divFn(args []interface{}) (value interface{}, err error) {
	if len(args) < 2 {
		return nil, fmt.Errorf(`function "/" takes two or more arguments`)
	}
	var resi int64
	var resf float64
	var f bool
	for i, arg := range args {
		switch arg := arg.(type) {
		case int64:
			if i == 0 && len(args) > 1 {
				resi = arg
				resf = float64(arg)
			} else {
				resi /= arg
				resf /= float64(arg)
			}
		case float64:
			if i == 0 && len(args) > 1 {
				resf = float64(arg)
			} else {
				resf /= arg
			}
			f = true
		default:
			return nil, fmt.Errorf("cannot divide with %#v", arg)
		}
	}
	if f {
		return resf, nil
	}
	return resi, nil
}

func andFn(scope *Scope, args []Node) (value interface{}, err error) {
	if len(args) == 0 {
		return true, nil
	}
	for _, arg := range args {
		value, err = scope.Eval(arg)
		if err != nil {
			return nil, err
		}
		if value == false {
			return false, nil
		}
	}
	return value, err
}

func orFn(scope *Scope, args []Node) (value interface{}, err error) {
	if len(args) == 0 {
		return false, nil
	}
	for _, arg := range args {
		value, err = scope.Eval(arg)
		if err != nil {
			return nil, err
		}
		if value != false {
			return value, nil
		}
	}
	return value, err
}

func ifFn(scope *Scope, args []Node) (value interface{}, err error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, errors.New(`function "if" takes two or three arguments`)
	}
	value, err = scope.Eval(args[0])
	if err != nil {
		return nil, err
	}
	if value == false {
		if len(args) == 3 {
			return scope.Eval(args[2])
		}
		return false, nil
	}
	return scope.Eval(args[1])
}

func varFn(scope *Scope, args []Node) (value interface{}, err error) {
	if len(args) == 0 || len(args) > 2 {
		return nil, errors.New("var takes one or two arguments")
	}
	symbol, ok := args[0].(*Symbol)
	if !ok {
		return nil, errors.New("var takes a symbol as first argument")
	}
	if len(args) == 1 {
		value = nil
	} else {
		value, err = scope.Eval(args[1])
		if err != nil {
			return nil, err
		}
	}
	return nil, scope.Create(symbol.Name, value)
}

func setFn(scope *Scope, args []Node) (value interface{}, err error) {
	if len(args) != 2 {
		return nil, errors.New(`function "set" takes two arguments`)
	}
	symbol, ok := args[0].(*Symbol)
	if !ok {
		return nil, errors.New(`function "set" takes a symbol as first argument`)
	}
	value, err = scope.Eval(args[1])
	if err != nil {
		return nil, err
	}
	return nil, scope.Set(symbol.Name, value)
}

func doFn(scope *Scope, args []Node) (value interface{}, err error) {
	scope = scope.Branch()
	for _, arg := range args {
		value, err = scope.Eval(arg)
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

func funcFn(scope *Scope, args []Node) (value interface{}, err error) {
	if len(args) < 2 {
		return nil, errors.New(`func takes three or more arguments`)
	}
	i := 0
	var name string
	if symbol, ok := args[0].(*Symbol); ok {
		name = symbol.Name
		i++
	}
	list, ok := args[i].(*List)
	if !ok {
		return nil, errors.New(`func takes a list of parameters`)
	}
	params := list.Nodes
	for _, param := range params {
		if _, ok := param.(*Symbol); !ok {
			return nil, errors.New("func's list of parameters must be a list of symbols")
		}
	}
	body := args[i+1:]
	if len(body) == 0 {
		return nil, fmt.Errorf("func takes a body sequence")
	}
	fn := func(args []interface{}) (value interface{}, err error) {
		if len(args) != len(params) {
			nameInfo := "anonymous function"
			if name != "" {
				nameInfo = fmt.Sprintf("function %q", name)
			}
			switch len(params) {
			case 0:
				return nil, fmt.Errorf("%s takes no arguments", nameInfo)
			case 1:
				return nil, fmt.Errorf("%s takes one argument", nameInfo)
			default:
				return nil, fmt.Errorf("%s takes %d arguments", nameInfo, len(params))
			}
		}
		scope = scope.Branch()
		for i, arg := range args {
			err := scope.Create(params[i].(*Symbol).Name, arg)
			if err != nil {
				panic("must not happen: " + err.Error())
			}
		}
		for _, node := range body {
			value, err = scope.Eval(node)
			if err != nil {
				return nil, err
			}
		}
		return value, nil
	}
	if name != "" {
		if err = scope.Create(name, fn); err != nil {
			return nil, err
		}
	}
	return fn, nil
}

func forFn(args []interface{}) (value interface{}, err error) {
	if len(args) != 2 {
		return nil, errors.New(`function "for" expects two arguments`)
	}
	fn, ok := args[1].(func([]interface{}) (interface{}, error))
	if !ok {
		return nil, fmt.Errorf(`function "for" expects function as second argument; got %T`, args[0])
	}
	switch iter := args[0].(type) {
	case []interface{}:
		pair := make([]interface{}, 2)
		for i, v := range iter {
			pair[0] = i
			pair[1] = v
			value, err = fn(pair)
			if err != nil {
				return nil, err
			}
		}
		return value, nil
	}
	return nil, fmt.Errorf(`function "for" cannot iterate over %T`, args[0])
}
