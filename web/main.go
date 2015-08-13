package main

import . "github.com/strickyak/basic_basic"
import "github.com/strickyak/basic_basic/draw"

import (
	"flag"
	. "fmt"
	//"io/ioutil"
	"log"
	"net/http"
	//"os"
)

func main() {
	flag.BoolVar(&Debug, "d", false, "debug bit")
	flag.Parse()

	http.HandleFunc("/", Render)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func Render(w http.ResponseWriter, req *http.Request) {
	defer func() {
		r := recover()
		if r != nil {
			w.Header().Set("Content-Type", "text/plain")
			Fprintf(w, "%v", r)
		}
	}()

	var code string
	req.ParseForm()
	if x, ok := req.Form["code"]; ok {
		code = x[0]

		terp := NewTerp(code)
		terp.SetExpiration("30s")
		d := draw.Register(terp)
		terp.Run()
		Printf("\n")
		w.Header().Set("Content-Type", "image/png")
		if d.HasImage() {
			d.WritePng(w)
		}
	} else {
		Fprintf(w, `
      <html><body>

        <form method="GET" action="/">
          <textarea name=code cols=80 rows=25>
1  REM Draw big grey triangle, then many smaller colored ones.
5  CALL triangle( 0,0, 0,99, 99,0, 444 )
10 for i = 0 to 9
20   for j = 0 to 9
30     for k = 0 to 9
40       let kk = 9 - k
44       call triangle (i*10,k+j*10,  9+i*10,j*10,  9+i*10,9+j*10, i+j*10+kk*100)
70     next k
80   next j
90 next i
          </textarea>
          <input type=submit name=submit value=Submit>
        </form>

<p><br><br><br>
<pre>
This is a simple BASIC computer.

The only data type is floating point numbers.

THe only output is the "CALL triangle" statement,
which draws colored triangles on a canvas with
coordinates [0 .. 100) on both x and y axes.

Statement ::= LineNumber Stmt
Stmt := REM remark...
  | LET var := expr
  | GOTO n
  | IF expr THEN y
  | IF expr THEN y ELSE n
  | FOR var = a TO b
  | NEXT var
  | GOSUB n
  | RETURN
  | CALL triangle( x1, y1, x2, y2, x3, y3, rgb )
       ... where n & y are line numbers
       ... where rgb is decimal (r=hundreds, g=tens, b=ones)
expr ::= sum relop expr     ...where relop can be == != < > <= >=
sum ::= prod addop sum      ...where addop can be + -
prod ::= prim mulop prod    ...where mulop can be * / %%
prim ::= number
  | var
  | ( expr )
</pre>
      </body></html>`)
	}
}
