package main

import (
	"bufio"
	"fmt"
	"os"

	"podojjit/interp"
	"podojjit/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide file name!")
		return
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Can't read file:", os.Args[1])
		os.Exit(1)
	}
	inp := bufio.NewReader(f)

	p := parser.NewParser().Parse(inp)
	interpreter := interp.NewInterpreter(p, 30000)
	err = interpreter.Interpret(true)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
