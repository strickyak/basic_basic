package basic_test

import . "github.com/strickyak/basic_basic"

import (
	"fmt"
	"testing"
)

func TestOne(t *testing.T) {
	Debug = true
	p1 := `10 let x = 42
         20 print x + 1`
	terp := NewTerp(p1)
	terp.Run()
	fmt.Printf("\n")
}
