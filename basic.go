/*
10 let n = 1000
20 for i = 1 to n
30 print i,
40 next i
50 stop
*/
package basic

import (
	`fmt`
	`regexp`
	`sort`
	`strconv`
	`strings`
)

var Debug bool

var F = fmt.Sprintf

var FindKeyword = regexp.MustCompile(`^(rem|let|print|if|then|else|for|next|stop)\b`).FindString
var FindNewline = regexp.MustCompile("^[;\n]").FindString
var FindWhite = regexp.MustCompile("^[ \t]*").FindString
var FindVar = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_*]*`).FindString
var FindNumber = regexp.MustCompile(`^[0-9]+`).FindString
var FindOp = regexp.MustCompile(`^[-+*/<=>]{1,2}`).FindString

type Kind int

const (
	EOF Kind = iota
	Keyword
	Var
	Number
	Newline
	Op
)

func (k Kind) String() string {
	switch k {
	case EOF:
		return "EOF"
	case Keyword:
		return "Keyword"
	case Var:
		return "Var"
	case Number:
		return "Number"
	case Newline:
		return "Newline"
	case Op:
		return "Op"
	default:
		return "?"
	}
}

type ForFrame struct {
	Var   string
	Value float64
	Max   float64
}

type Line struct {
	N   int
	Cmd Cmd
}
type LineSlice []Line                  // Can sort.
func (o LineSlice) Len() int           { return len(o) }
func (o LineSlice) Less(i, j int) bool { return o[i].N < o[j].N }
func (o LineSlice) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

type Terp struct {
	// Lexer state.
	Program string
	P       int
	K       Kind
	S       string
	F       float64
	// Interpreter state.
	G           map[string]float64
	Stack       []*ForFrame
	Lines       LineSlice
	LastPrinted float64
}

func NewTerp(program string) *Terp {
	t := &Terp{
		Program: program,
		G:       make(map[string]float64),
	}
	t.Lines = t.ParseProgram()
	sort.Sort(t.Lines)
	return t
}

func (t *Terp) Run() {
	t.G = make(map[string]float64)
	t.Stack = make([]*ForFrame, 0)
	i := 0
	n := len(t.Lines)
	for i < n {
		j := t.Lines[i].Cmd.Eval(t)
		if j > 0 {
			for i := 0; i < n; i++ {
				if t.Lines[i].N == j {
					break
				}
			}
		} else {
			i++
		}
	}
}

func (o *Terp) Advance() {
	o.Advance1()
	if Debug {
		println(F("[[ %5d:   %10v %10q ... %q ]]", o.P, o.K, o.S, o.Program[o.P:]))
	}
}
func (o *Terp) Advance1() {
	m := FindWhite(o.Program[o.P:])
	o.P += len(m)
	if o.P == len(o.Program) {
		o.K = EOF
		o.S = ""
		return
	}
	m = FindNewline(o.Program[o.P:])
	if m != "" {
		o.K = Newline
		o.S = m
		o.P += len(m)
		return
	}
	m = FindKeyword(o.Program[o.P:])
	if m != "" {
		o.K = Keyword
		o.S = m
		o.P += len(m)
		return
	}
	m = FindVar(o.Program[o.P:])
	if m != "" {
		o.K = Var
		o.S = m
		o.P += len(m)
		return
	}
	m = FindNumber(o.Program[o.P:])
	if m != "" {
		o.K = Number
		o.S = m
		o.F, _ = strconv.ParseFloat(m, 64)
		o.P += len(m)
		return
	}
	m = FindOp(o.Program[o.P:])
	if m != "" {
		o.K = Op
		o.S = m
		o.P += len(m)
		return
	}
	panic(F("Cannot Lex: 5q", o.Program[o.P:]))
}

type Expr struct {
	Const *float64
	Var   *string
	A, B  *Expr
	Op    string
}

func (o *Expr) Eval(t *Terp) float64 {
	switch {
	case o.Const != nil:
		return *o.Const
	case o.Var != nil:
		return t.G[*o.Var]
	case o.Op != "":
		a := o.A.Eval(t)
		b := o.B.Eval(t)
		switch o.Op {
		case "+":
			return a + b
		default:
			panic(F("bad op: %q", o.Op))
		}
	default:
		panic("bad expr")
	}
}

type NopCmd struct{}

func (o *NopCmd) Show(t *Terp) string { return "Nop..." }
func (o *NopCmd) Eval(t *Terp) int {
	return 0
}

type PrintCmd struct {
	X *Expr
}

func (o *PrintCmd) Show(t *Terp) string { return "Print..." }
func (o *PrintCmd) Eval(t *Terp) int {
	t.LastPrinted = o.X.Eval(t)
	if Debug {
		println(F("## PRINT %g", t.LastPrinted))
	}
	fmt.Printf("%g ", t.LastPrinted)
	return 0
}

type LetCmd struct {
	Var string
	X   *Expr
}

func (o *LetCmd) Eval(t *Terp) int {
	z := o.X.Eval(t)
	if Debug {
		println(F("## LET %s = %g", o.Var, z))
	}
	t.G[o.Var] = z
	return 0
}
func (o *LetCmd) Show(t *Terp) string { return "Let..." }

type ForCmd struct {
	Var   string
	Value *Expr
	Max   *Expr
}
type NextCmd struct {
	Var string
}
type Cmd interface {
	Eval(t *Terp) int
	Show(t *Terp) string
}

func (lex *Terp) ParsePrim() *Expr {
	z := &Expr{}
	switch lex.K {
	case Var:
		s := lex.S
		lex.Advance()
		z.Var = &s
	case Number:
		f := lex.F
		lex.Advance()
		z.Const = &f
	default:
		panic("expected prim")
	}
	return z
}
func (lex *Terp) ParseExpr() *Expr {
	a := lex.ParsePrim()
	for lex.K == Op {
		op := lex.S
		lex.Advance()
		b := lex.ParsePrim()
		a = &Expr{A: a, Op: op, B: b}
	}
	return a
}

func (lex *Terp) ParseProgram() []Line {
	var c Cmd
	var w string // the command
	var n int    // the line number
	var z []Line
Loop:
	for {
		lex.Advance()
		// Want line number.
		switch lex.K {
		case EOF:
			break Loop
		case Newline:
			continue
		case Number:
			n = int(lex.F)
		default:
			panic("expected line number")
		}
		lex.Advance()
		// Want command.
		switch lex.K {
		case Keyword:
			w = lex.S
		case Newline:
			w = ";"
		default:
			panic("expected command")
		}
		lex.Advance()
		switch strings.ToLower(w) {
		case ";":
			c = &NopCmd{}
		case "rem":
			c = &NopCmd{}
			for lex.K != EOF && lex.K != Newline {
				lex.Advance()
			}
		case "print":
			x := lex.ParseExpr()
			c = &PrintCmd{X: x}
		case "let":
			Check(lex.K == Var, "expected var after let")
			v := lex.S
			lex.Advance()
			Check(lex.K == Op && lex.S == "=", "expected = after var after let")
			lex.Advance()
			x := lex.ParseExpr()
			c = &LetCmd{Var: v, X: x}
		default:
			panic("unknown command: " + w)
		}
		z = append(z, Line{N: n, Cmd: c})
	}
	return z
}

func Check(b bool, s string) {
	if !b {
		panic(s)
	}
}
