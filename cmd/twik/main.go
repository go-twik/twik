package main

import (
	"launchpad.net/twik"
	"code.google.com/p/go.crypto/ssh/terminal"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printfFn(args []interface{}) (interface{}, error) {
	if len(args) > 0 {
		if format, ok := args[0].(string); ok {
			_, err := fmt.Printf(format, args[1:]...)
			return nil, err
		}
	}
	return nil, fmt.Errorf("printf takes a format string")
}

func listFn(args []interface{}) (interface{}, error) {
	return args, nil
}

func run() error {
	fset := twik.NewFileSet()
	scope := twik.NewScope(fset)
	scope.Create("printf", printfFn)
	scope.Create("list", listFn)

	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "-") {
			return fmt.Errorf("usage: twik [<source file>]")
		}
		f, err := os.Open(os.Args[1])
		if err != nil {
			return err
		}
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		node, err := twik.Parse(fset, os.Args[1], data)
		if err != nil {
			return err
		}

		_, err = scope.Eval(node)
		return err
	}

	state, err := terminal.MakeRaw(1)
	if err != nil {
		return err
	}
	defer terminal.Restore(1, state)

	t := terminal.NewTerminal(os.Stdout, "> ")
	unclosed := ""
	for {
		line, err := t.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if unclosed != "" {
			line = unclosed + "\n" + line
		}
		unclosed = ""
		t.SetPrompt("> ")
		node, err := twik.ParseString(fset, "", line)
		if err != nil {
			if strings.HasSuffix(err.Error(), "missing )") {
				unclosed = line
				t.SetPrompt(". ")
				continue
			}
			fmt.Println(err)
			continue
		}
		value, err := scope.Eval(node)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if value != nil {
			if reflect.TypeOf(value).Kind() == reflect.Func {
				fmt.Println("#func")
			} else if v, ok := value.([]interface{}); ok {
				if len(v) == 0 {
					fmt.Println("()")
				} else {
					fmt.Print("(list")
					for _, e := range v {
						fmt.Printf(" %#v", e)
					}
					fmt.Println(")")
				}
			} else {
				fmt.Printf("%#v\n", value)
			}
		}
	}
	fmt.Println()
	return nil
}
