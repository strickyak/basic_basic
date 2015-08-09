package basic_test

import . "github.com/strickyak/basic_basic"

import (
	"fmt"
	"testing"
)

type PF struct {
	Prog  string
	Final float64
}

func TestOne(t *testing.T) {
	progs := []PF{
		PF{`  10 print 3*3+4*4`, 25},
		PF{`  10 let x = 42
          20 print x + 1`, 43},
		PF{`44 goto 88
        55 print 13
        88 print 99`, 99},
		PF{`100 print 4 == 5`, 0},
		PF{`100 print 4 == 4`, 1},
		PF{`10 for i = 1 to 5
        20 print i + 10
        30 next i`, 15},
		PF{`5 let s = 0;
        10 for i = 1 to 10;
        20 let s = s + i;
        30 next i;
        99 print s`, 55},
		PF{`5 REM gosub and return
        10 let x = 5
        20 gosub 200
        30 let x = x + 3
        40 gosub 300
        90 goto 9999
        200 rem double the number
        210 let x = x * 2
        220 return
        300 rem sextuple the number
        310 let x = x * 3
        320 gosub 200
        330 return
        9999 print x`, 78},
		PF{`5 REM Fizz Buzz
        10 for i = 1 to 100
        20 if i % 15 == 0 then 400
        30 if i % 3 == 0 then 300
        40 if i % 5 == 0 then 500
        50 print i
        60 goto 900
        300 print -3
        310 goto 900
        400 print -15
        410 goto 900
        500 print -5
        510 goto 900
        900 next i`, -5},
	}
	Debug = true
	for i, pf := range progs {
		t.Logf("Test Program %d: %q", i, pf.Prog)
		terp := NewTerp(pf.Prog)
		terp.Run()
		fmt.Printf("\n")
		final := terp.LastPrinted
		if final != pf.Final {
			t.Errorf("bad final: got %g want %g", final, pf.Final)
		}
	}
}
