/*
  Old School BASIC.  Only numbers, so far.
*/
package basic

import (
	`fmt`
	//`os`
	`regexp`
	`sort`
	`strconv`
	`strings`
)

var Debug bool
var Epsilon = 0.000001

var F = fmt.Sprintf

var FindKeyword = regexp.MustCompile(`^(?i)(rem|let|dim|print|goto|gosub|return|if|then|else|for|to|next|stop|call)\b`).FindString
var FindNewline = regexp.MustCompile("^[;\n]").FindString
var FindWhite = regexp.MustCompile("^[ \t]*").FindString
var FindVar = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_*]*`).FindString
var FindNumber = regexp.MustCompile(`^-?[0-9]+[.]?[0-9]*`).FindString
var FindPunc = regexp.MustCompile(`^[()[]{},;]`).FindString
var FindOp = regexp.MustCompile(`^[^A-Za-z0-9;\s]+`).FindString

type Kind int

const (
	EOF Kind = iota
	Keyword
	Var
	Number
	Newline
	Punc
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

type Line struct {
	N   int
	Cmd Cmd
}
type LineSlice []Line                  // Can sort.
func (o LineSlice) Len() int           { return len(o) }
func (o LineSlice) Less(i, j int) bool { return o[i].N < o[j].N }
func (o LineSlice) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

type Callable func(t *Terp, args []float64) float64

func (t *Terp) AddExtension(name string, fn Callable)  {
  name = strings.ToLower(name)
  t.Extensions[name] = fn
}

type Terp struct {
	// Lexer state.
	Program string
	P       int
	K       Kind
	S       string
	F       float64
	// Interpreter state.
	G           map[string]float64
	ForFrames   []*ForFrame
	GosubFrames []*GosubFrame
	Lines       LineSlice
	Line        int
	LastPrinted float64
  Extensions map[string]Callable
}

func NewTerp(program string) *Terp {
	t := &Terp{
		Program: program,
		G:       make(map[string]float64),
    Extensions: make(map[string]Callable),
	}
	println(F("NewTerp-- P=%d K=%v S=%s", t.P, t.K, t.S))
	t.Advance()
	println(F("NewTerp-- P=%d K=%v S=%s", t.P, t.K, t.S))
	t.Lines = t.ParseProgram()
	println(F("NewTerp-- P=%d K=%v S=%s", t.P, t.K, t.S))
	sort.Sort(t.Lines)
	return t
}

func (t *Terp) Run() {
	t.G = make(map[string]float64)
	t.ForFrames = make([]*ForFrame, 0)
	i := 0
	n := len(t.Lines)
Loop:
	for i < n {
		t.Line = t.Lines[i].N
		println(F("%d: eval<< %v", i, t.Lines[i]))
		j := t.Lines[i].Cmd.Eval(t)
		println(F("%d: eval>>%d", i, j))

		if j > 0 {
			// Branching instruction.
			for i = 0; i < n; i++ {
				println(F("%d: look got %d want %d", i, t.Lines[i].N, j))
				if t.Lines[i].N >= j {
					println(F("%d: CONTINUE got %d want %d", i, t.Lines[i].N, j))
					continue Loop
				}
			}
			println(F("FALLOUT"))
		} else {
			// Proceed to next instrution.
			i++
			println(F("NEXT"))
		}
	}
	println(F("END"))
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
	panic(F("Cannot Lex: %q", o.Program[o.P:]))
}

type Expr struct {
	Const *float64
	Var   *string
	A, B  *Expr
	Op    string
}

func (o *Expr) String() string {
	if o.Const != nil {
		return F("(Const:%g)", *o.Const)
	}
	if o.Var != nil {
		return F("(Var:%s)", *o.Var)
	}
	return F("(%s %s %s)", o.A.String(), Op, o.B.String())
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
		case "-":
			return a - b
		case "*":
			return a * b
		case "/":
			return a / b
		case "%":
			return float64(TrimInt(a) % TrimInt(b))
		case "==":
			return Truth(a == b)
		default:
			panic(F("bad op: %q", o.Op))
		}
	default:
		panic("bad expr")
	}
}

func TrimInt(f float64) int {
	println(F("TrimInt <<< %g", f))
	h := 0.5
	if f < 0 {
		h = -0.5
	}
	i := int(f + h)
	d := f - float64(i)
	println(F("TrimInt >>> %i +- %f", i, d))
	if d < -Epsilon || d > Epsilon {
		panic(F("Not an integer: %g", f))
	}
	return i
}

type NopCmd struct{}

func (o *NopCmd) String() string { return "Nop..." }
func (o *NopCmd) Eval(t *Terp) int {
	return 0
}

type GosubFrame struct {
	ReturnLine int
}
type GosubCmd struct {
	CallLine int
}

func (o *GosubCmd) String() string { return F("Gosub %d", o.CallLine) }
func (o *GosubCmd) Eval(t *Terp) int {
	fr := &GosubFrame{t.Line + 1} // Return Address.
	t.GosubFrames = append(t.GosubFrames, fr)
	return o.CallLine // Call address.
}

type ReturnCmd struct {
	N int
}

func (o *ReturnCmd) String() string { return F("Return") }
func (o *ReturnCmd) Eval(t *Terp) int {
	n := len(t.GosubFrames)
	fr := t.GosubFrames[n-1]
	t.GosubFrames = t.GosubFrames[:n-1]
	return fr.ReturnLine
}

type GotoCmd struct {
	N int
}

func (o *GotoCmd) String() string { return F("Goto %d", o.N) }
func (o *GotoCmd) Eval(t *Terp) int {
	return o.N
}

type PrintCmd struct {
	X *Expr
}

func (o *PrintCmd) String() string { return F("Print %v", o.X) }
func (o *PrintCmd) Eval(t *Terp) int {
	t.LastPrinted = o.X.Eval(t)
	if Debug {
		println(F("## PRINT %g", t.LastPrinted))
	}
	fmt.Printf("%g ", t.LastPrinted)
	return 0
}

type CallCmd struct {
  Var string
  Args []*Expr
}
func (o *CallCmd) String() string { return F("CALL %s (%v)", o.Var, o.Args) }
func (o *CallCmd) Eval(t *Terp) int {
  ext, ok := t.Extensions[o.Var]
  if !ok {
    panic(F("cannot call unknown extension: %q", o.Var))
  }
  var args []float64
  for _, e := range o.Args {
    args = append(args, e.Eval(t))
  }
  _ = ext(t, args)
  return 0
}

type LetCmd struct {
	Var string
	X   *Expr
}

func (o *LetCmd) Eval(t *Terp) int {
	x := o.X.Eval(t)
	if Debug {
		println(F("## LET %s = %g", o.Var, x))
	}
	t.G[o.Var] = x
	return 0
}
func (o *LetCmd) String() string { return F("LET %s = %v", o.Var, o.X) }

type ForFrame struct {
	Var   string
	Value float64
	Max   float64
	Line  int
}
type NextCmd struct {
	Var string
}

func (o *NextCmd) String() string { return F("NEXT %s", o.Var) }
func (o *NextCmd) Eval(t *Terp) int {
	var i int
	for i = len(t.ForFrames) - 1; i >= 0; i-- {
		if t.ForFrames[i].Var == o.Var {
			break
		}
	}
	if i < 0 {
		panic(F("NEXT %s: cannot find FOR frame", o.Var))
	}
	fr := t.ForFrames[i]
	fr.Value += 1
	t.G[o.Var] = fr.Value
	if fr.Value > fr.Max {
		t.ForFrames = t.ForFrames[:i]
		return 0
	}
	t.ForFrames = t.ForFrames[:i+1]
	return fr.Line
}

type ForCmd struct {
	Var   string
	Begin *Expr
	End   *Expr
}

func (o *ForCmd) String() string { return F("FOR %s = %v TO %v", o.Var, o.Begin, o.End) }
func (o *ForCmd) Eval(t *Terp) int {
	begin := o.Begin.Eval(t)
	end := o.End.Eval(t)
	if Debug {
		println(F("## FOR %s = %g TO %g", o.Var, begin, end))
	}
	t.G[o.Var] = begin
	fr := &ForFrame{Var: o.Var, Value: begin, Max: end, Line: t.Line + 1}
	t.ForFrames = append(t.ForFrames, fr)
	return 0
}

type IfCmd struct {
	Cond *Expr
	Then int
	Else int
}

func (o *IfCmd) String() string { return F("IF %v THEN %d ELSE %d", o.Cond, o.Then, o.Else) }
func (o *IfCmd) Eval(t *Terp) int {
	cond := o.Cond.Eval(t)
	if TrimInt(cond) == 0 {
		return o.Else
	}
	return o.Then
}

type Cmd interface {
	Eval(t *Terp) int
	String() string
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
	for lex.K == Op && lex.S != ")" && lex.S != "," {
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
		for lex.K == Newline {
			lex.Advance()
		}
		println(F("LOOP %d %q", lex.P, lex.Program[lex.P:]))
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
		println(F("LINE %d", n))
		lex.Advance()
		// Want command.
		switch lex.K {
		case Keyword:
			w = lex.S
			lex.Advance()
		case Newline:
			w = ";"
		default:
			panic("expected command after line number")
		}
		println(F("KEYWORD %s", w))
		switch strings.ToLower(w) {
		case ";":
			c = &NopCmd{}
		case "rem":
			c = &NopCmd{}
			for lex.K != EOF && lex.K != Newline {
				lex.Advance()
			}
		case "goto":
			n := int(lex.ParseNumber())
			c = &GotoCmd{N: n}
		case "gosub":
			n := int(lex.ParseNumber())
			c = &GosubCmd{CallLine: n}
		case "return":
			c = &ReturnCmd{}
		case "print":
			x := lex.ParseExpr()
			c = &PrintCmd{X: x}
		case "let":
			v := lex.ParseVar()
			lex.ParseMustSym("=")
			x := lex.ParseExpr()
			c = &LetCmd{Var: v, X: x}
		case "next":
			v := lex.ParseVar()
			c = &NextCmd{Var: v}
		case "for":
			v := lex.ParseVar()
			lex.ParseMustSym("=")
			x := lex.ParseExpr()
			lex.ParseMustKeyword("to")
			y := lex.ParseExpr()
			c = &ForCmd{Var: v, Begin: x, End: y}
		case "if":
			cond := lex.ParseExpr()
			lex.ParseMustKeyword("then")
			ifTrue := TrimInt(lex.ParseNumber())
			ifFalse := 0
			if lex.S == "else" {
				lex.ParseMustKeyword("else")
				ifFalse = TrimInt(lex.ParseNumber())
			}
			c = &IfCmd{Cond: cond, Then: ifTrue, Else: ifFalse}
    case "call":
			name := strings.ToLower(lex.ParseVar())
      println("name=", name)
			lex.ParseMustSym("(")
      var args []*Expr
      for {
        for lex.S == "," {
          println("comma")
          lex.Advance()
        }
        if lex.S == ")" { break }
        println(" not close paren ", lex.S)
			  a := lex.ParseExpr()
        println("  got arg ", a)
        args = append(args, a)
      }
			lex.ParseMustSym(")")
      c = &CallCmd{name, args}
		default:
			panic("unknown command: " + w)
		}
		z = append(z, Line{N: n, Cmd: c})
		if lex.K == EOF {
			break
		}
		if lex.K != Newline {
			panic("extra stuff at end of command")
		}
	}
	return z
}
func (lex *Terp) ParseVar() string {
	Check(lex.K == Var, "expected variable name")
	s := lex.S
	lex.Advance()
	return s
}
func (lex *Terp) ParseNumber() float64 {
	Check(lex.K == Number, "expected variable name")
	f := lex.F
	lex.Advance()
	return f
}
func (lex *Terp) ParseMustSym(x string) {
	Check(lex.K == Op && lex.S == x, "expected symbol: "+x)
	lex.Advance()
}
func (lex *Terp) ParseMustKeyword(x string) {
	Check(lex.K == Keyword && lex.S == x, "expected keyword: "+x)
	lex.Advance()
}

func Truth(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func Check(b bool, s string) {
	if !b {
		panic(s)
	}
}
