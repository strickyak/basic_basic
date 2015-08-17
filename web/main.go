package main

import . "github.com/strickyak/basic_basic"
import "github.com/strickyak/basic_basic/draw"

import (
	"bufio"
	"bytes"
	"flag"
	. "fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var Tmpl *template.Template

var NON_ALPHANUM = regexp.MustCompile("[^A-Za-z0-9_]")

func encodeChar(s string) string {
	w := bytes.NewBuffer(nil)
	r := strings.NewReader(s)
	for r.Len() > 0 {
		ch, _, err := r.ReadRune()
		if err != nil {
			panic(err)
		}
		Fprintf(w, "{%d}", ch)
	}
	return w.String()
}
func CurlyEncode(s string) string {
	if s == "" {
		return "{}" // Special case for encoding the empty string.
	}
	return NON_ALPHANUM.ReplaceAllStringFunc(s, encodeChar)
}

func main() {
	flag.BoolVar(&Debug, "d", false, "debug bit")
	flag.Parse()

	Tmpl = template.New("basic-web")
	Tmpl.Parse(TEMPLATES)

	http.HandleFunc("/", handler)
	log.Println("http.ListenAndServe.")
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		r := recover()
		if r != nil {
			w.Header().Set("Content-Type", "text/plain")
			Fprintf(w, "%v", r)
		}
	}()

	req.ParseForm()
	if f4, ok4 := req.Form["list"]; ok4 {
		what := f4[0]
		if what == "" {
			// List all files
			names, err := filepath.Glob("*.bas")
			if err != nil {
				panic(err)
			}
			sort.Strings(names)
			w.Header().Set("Content-Type", "text/html")
			Fprintf(w, "<html><body>")
			for _, name := range names {
				Fprintf(w, `<a href="/?list=%s">%s</a><br>`+"\n", name, name)
			}
		} else {
			// List one file
			fd, err := os.Open(what)
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Type", "text/plain")
			_, err = io.Copy(w, fd)
			if err != nil {
				panic(err)
			}
			err = fd.Close()
			if err != nil {
				panic(err)
			}
		}
	} else if f1, ok1 := req.Form["run"]; ok1 {
		var putchar func(ch byte)
		forward_putchar := func(ch byte) {
			putchar(ch)
		}

		terp := NewTerp(f1[0], forward_putchar)
		terp.SetExpiration("30s")
		d := draw.Register(terp)
		putchar = d.Putchar

		if f3, ok3 := req.Form["progname"]; ok3 {
			if f3[0] == "" {
				f3[0] = "untitled"
			}
			fd, err := os.OpenFile(CurlyEncode(f3[0])+".bas", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				panic(err)
			}
			w := bufio.NewWriter(fd)
			Fprintf(w, "###### ###### ###### ###### ###### ######\n")
			Fprintf(w, "%s\n", strings.Replace(f1[0], "\r", "", -1))
			Fprintf(w, "###### ###### ###### ###### ###### ######\n")
			w.Flush()
			fd.Close()
		}

		terp.Run()
		if d.HasImage() {
			w.Header().Set("Content-Type", "image/png")
			d.WritePng(w)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			Fprintf(w, "Use 'PRINT' or 'CALL triangle' statements to produce output.")
		}
	} else {
		dict := make(map[string]interface{})
		if f2, ok2 := req.Form["load"]; ok2 {
			dict["Code"] = f2[0]
		} else {
			dict["Code"] = template.HTML(DEMO)
		}
		Tmpl.ExecuteTemplate(w, "Main", dict)
	}
}

const TEMPLATES = `
{{define "Main"}}
    <html><body>
      <form method="POST" action="/">
        <textarea name=run cols=80 rows=25>{{$.Code}}</textarea>
        <br>
        <input type=submit name=submit value=Submit>
        &nbsp; &nbsp; &nbsp; &nbsp;
        ( Save as: <input type=text width=20 name=progname> )
      </form>
      {{template "Doc" $}}
{{end}}

{{define "Doc"}}
<p>
<a href="/?list=">See saved programs.</a>
<p>
<pre>
This is a simple BASIC computer.

The only data type is floating point numbers.

THe only output is the "CALL triangle" statement,
which draws colored triangles on a canvas with
coordinates [0 .. 100) on both x and y axes.

Statement ::= LineNumber Stmt
Stmt := REM remark...
  | DIM arr(size), matrix(width,heigth)
  | LET var := expr
  | LET arr(i, j...) := expr
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
prod ::= composite mulop prod    ...where mulop can be * / %%
composite ::= prim
  | arr(i, j...)
prim ::= number
  | var
  | ( expr )
</pre>
      </body></html>
    </body></html>
{{end}}
`

const DEMO = `
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
`
