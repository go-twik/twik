package main

import (
	"launchpad.net/twik"
	"code.google.com/p/go.crypto/ssh/terminal"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	state, err := terminal.MakeRaw(1)
	if err != nil {
		return err
	}
	defer terminal.Restore(1, state)

	fset := twik.NewFileSet()
	scope := twik.NewScope(fset)

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
			fmt.Printf("%#v\n", value)
		}
	}
	fmt.Println()
	return nil
}
