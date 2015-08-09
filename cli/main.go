/*
$ echo '555 print 3 + 88;88 print 9; 44' | go run cli/main.go 2>/dev/null
9 91
*/
package main

import . "github.com/strickyak/basic_basic"
import "github.com/strickyak/basic_basic/draw"

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	prog, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	terp := NewTerp(string(prog))
  d := draw.Register(terp)
	terp.Run()
	fmt.Printf("\n")
  if d.HasImage() {
    d.SavePng("/tmp/out.png")
  }
}
