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
110 CALL triangle( 0,0, 0,99, 99,0, 909 )
          </textarea>
          <input type=submit name=submit value=Submit>
        </form>

      </body></html>`)
	}
}
