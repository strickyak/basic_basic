/*
$ echo '555 print 3 + 88;88 print 9; 44' | go run cli/main.go 2>/dev/null
9 91
*/
package main

import . "github.com/strickyak/basic_basic"
import "github.com/strickyak/basic_basic/draw"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func putchar(ch byte) {
  fmt.Printf("<%c>", ch)
}

func main() {
	flag.BoolVar(&Debug, "d", false, "debug bit")
	var filename string
	flag.StringVar(&filename, "f", "", "basic source file")
	flag.Parse()

	var prog []byte
	var err error
	if filename == "" {
		prog, err = ioutil.ReadAll(os.Stdin)
	} else {
		prog, err = ioutil.ReadFile(filename)
	}
	if err != nil {
		panic(err)
	}
	terp := NewTerp(string(prog), putchar)
	d := draw.Register(terp)
	terp.Run()
	fmt.Printf("\n")
	if d.HasImage() {
		d.SavePng("/tmp/out.png")
	}
}
