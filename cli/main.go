package main

import . "github.com/strickyak/basic_basic"

import (
	"fmt"
)

func main() {
	p1 := `10 let x = 42
         20 print x + 1`
	lines := ParseProgram(p1)
	terp := NewTerp()
	for _, e := range lines {
		e.Cmd.Run(terp)
	}
	fmt.Printf("\n")
}
