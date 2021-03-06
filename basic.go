/*
  Old School BASIC.  Only numbers, so far.
*/
package basic

import (
	`fmt`
	`log`
	`math`
	`os`
	`regexp`
	`runtime/debug`
	`sort`
	`strconv`
	`strings`
	`time`
)

const Epsilon = 0.000001 // For snapping to integer.
const RecursionLimit = 1000

var Debug bool

var F = fmt.Sprintf

var FindKeyword = regexp.MustCompile(`^(?i)(REM|LET|DIM|PRINT|GOTO|GOSUB|RETURN|IF|THEN|ELSE|FOR|TO|NEXT|STOP|CALL)\b`).FindString

var FindNewline = regexp.MustCompile("^[;\n]").FindString                                    // Semicolons are newlines.
var FindWhite = regexp.MustCompile("^[ \t\r]*").FindString                                   // But not newlines.
var FindVar = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*`).FindString                       // Like C identifiers.
var FindNumber = regexp.MustCompile(`^-?[0-9]+[.]?[0-9]*`).FindString                        // Not yet E notaion.
var FindPunc = regexp.MustCompile(`^[][(){},;]`).FindString                                  // Single char punc.
var FindOp = regexp.MustCompile(`^(==|!=|<=|>=|<>|[*][*]|<<|>>|[^A-Za-z0-9;\s])`).FindString // Contains some double-char sequences.
var FindString = regexp.MustCompile(`^"([^"]|"")*"`).FindString                              // String Literal.

type Kind int

const (
	EOF Kind = iota
	Keyword
	Var
	Number
	String
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
	case String:
		return "String"
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

func (t *Terp) AddExtension(name string, fn Callable) {
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
	Scalars     map[string]float64
	Arrays      map[string]interface{}
	ForFrames   []*ForFrame
	GosubFrames []*GosubFrame
	Lines       LineSlice
	Line        int
	LastPrinted float64
	Extensions  map[string]Callable
	Expiration  *time.Time
	Putchar     func(ch byte)
}

func NewTerp(program string, putchar func(ch byte)) *Terp {
	t := &Terp{
		Program:    program,
		Scalars:    make(map[string]float64),
		Arrays:     make(map[string]interface{}),
		Extensions: make(map[string]Callable),
		Putchar:    putchar,
	}
	t.Advance()
	t.Lines = t.ParseProgram()
	sort.Sort(t.Lines)
	return t
}

func (t *Terp) SetExpiration(duration string) {
	dur, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}
	z := time.Now().Add(dur)
	t.Expiration = &z
}

func (t *Terp) CheckExpiration() {
	if t.Expiration != nil && time.Now().After(*t.Expiration) {
		panic("TIMEOUT: Your program ran too long.")
	}
}

func AllocArray(dims []int) interface{} {
	switch len(dims) {
	case 0:
		panic("missing dim")
	case 1:
		d := dims[0]
		// Allocate 1 extra, cauz this is Basic.
		return make([]float64, d+1)
	default:
		d := dims[0]
		// Allocate 1 extra, cauz this is Basic.
		vec := make([]interface{}, d+1)
		for i := 0; i <= d; i++ {
			vec[i] = AllocArray(dims[1:])
		}
		return vec
	}
}

func (t *Terp) Run() {
	t.Scalars = make(map[string]float64)
	t.ForFrames = make([]*ForFrame, 0)
	i := 0
	n := len(t.Lines)

	// Allocate Arrays.
	for _, line := range t.Lines {
		if dim, ok := line.Cmd.(*DimCmd); ok {
			t.Arrays[dim.Var] = AllocArray(dim.Dims)
		}
	}

	if !Debug {
		defer func() {
			r := recover()
			if r != nil {
				log.Printf("\n\nERROR: %v\n", r)
				debug.PrintStack()
				log.Printf("\n\n", r)
				s := F("%v", r)
				if i >= 0 && i < len(t.Lines) {
					s = F("%s; in line number %d", s, t.Lines[i].N)
				}
				panic(s)
			}
		}()
	}

	gotoCache := make(map[int]int) // Line Num Dest -> Lines Slice Index
BasicLoop:
	for i < n {
		t.Line = t.Lines[i].N
		j := t.Lines[i].Cmd.Eval(t)

		switch {
		case j < 0:
			break BasicLoop // STOP and END commands.

		case j > 0:
			if index, ok := gotoCache[j]; ok {
				i = index
				continue BasicLoop
			}
			// Branching instruction.
			t.CheckExpiration()
			for i = 0; i < n; i++ {
				if t.Lines[i].N >= j {
					gotoCache[j] = i
					continue BasicLoop
				}
			}

		case j == 0:
			// Proceed to next instrution.
			i++
		}
	}
}

func (o *Terp) Advance() {
	o.Advance1()
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
		o.S = strings.ToUpper(m)
		o.P += len(m)
		return
	}
	m = FindVar(o.Program[o.P:])
	if m != "" {
		o.K = Var
		o.S = strings.ToUpper(m)
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
	m = FindString(o.Program[o.P:])
	if m != "" {
		o.K = String
		o.S = strings.Replace(m[1:len(m)-1], `""`, `"`, -1)
		o.P += len(m)
		return
	}
	m = FindPunc(o.Program[o.P:])
	if m != "" {
		o.K = Punc
		o.S = m
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
	Subs  []*Expr
}

func (o *Expr) String() string {
	if o.Const != nil {
		return F("(Const:%g)", *o.Const)
	}
	if o.Var != nil {
		return F("(Var:%s)", *o.Var)
	}
	return F("(%s {%s} %s)", o.A.String(), o.Op, o.B.String())
}

func (o *Expr) Eval(t *Terp) float64 {
	var z float64
	switch {
	case o.Subs != nil:
		if o.A.Var == nil {
			panic("Expected var name for subscripting")
		}
		name := *o.A.Var
		if name == "" {
			panic("Expected name of array")
		}

		thing, ok := t.Arrays[name]
		if !ok {
			panic("No such array: " + name)
		}

		subs := o.Subs
		for len(subs) > 0 {
			e := SnapToInt(subs[0].Eval(t))
			if e < 0 {
				panic(F("negative subscript: array %s got %d", name, e))
			}
			switch t := thing.(type) {
			case []interface{}:
				if len(subs) <= 1 {
					panic("too few subscripts for array: " + name)
				}
				if e > len(t) {
					panic(F("subscript too big: array %q got %d want <=%d", name, e, len(t)))
				}
				thing = t[e]
			case []float64:
				if len(subs) != 1 {
					panic("too many subscripts for array: " + name)
				}
				if e > len(t) {
					panic(F("subscript too big: array %q got %d want <=%d", name, e, len(t)))
				}
				return t[e]
			default:
			}
			subs = subs[1:]
		}
		panic("notreached")

	case o.Const != nil:
		z = *o.Const
	case o.Var != nil:
		z = t.Scalars[*o.Var]
	case o.Op != "":
		a := o.A.Eval(t)
		b := o.B.Eval(t)
		switch o.Op {
		case "+":
			z = a + b
		case "-":
			z = a - b
		case "*":
			z = a * b
		case "/":
			z = a / b
		case "%":
			z = float64(SnapToInt(a) % SnapToInt(b))
		case "==", "=":
			z = Truth(a == b)
		case "!=", "<>":
			z = Truth(a != b)
		case "<=":
			z = Truth(a <= b)
		case ">=":
			z = Truth(a >= b)
		case "<":
			z = Truth(a < b)
		case ">":
			z = Truth(a > b)
		default:
			panic(F("bad op: %q", o.Op))
		}
	default:
		panic("bad expr")
	}
	if math.IsInf(z, 0) {
		panic("Result is Infinity")
	}
	if math.IsNaN(z) {
		panic("Result is Not a Number")
	}
	return z
}

func SnapToInt(f float64) int {
	h := 0.5
	if f < 0 {
		h = -0.5
	}
	i := int(f + h)
	d := f - float64(i)
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
	if len(t.GosubFrames) >= RecursionLimit {
		panic(F("Recursion Limit: You called GOSUB too many times."))
	}
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

type StopCmd struct {
}
type EndCmd struct {
}

func (o *StopCmd) String() string { return F("Stop") }
func (o *StopCmd) Eval(t *Terp) int {
	return -1
}
func (o *EndCmd) String() string { return F("End") }
func (o *EndCmd) Eval(t *Terp) int {
	return -1
}

type PrintCmd struct {
	X *Expr  // union, if floating expr
	S string // union, if string literal
}

func (o *PrintCmd) String() string {
	if o.X == nil {
		return F("Print %q", o.S)
	} else {
		return F("Print %v", o.X)
	}
}
func (o *PrintCmd) Eval(t *Terp) int {
	if o.X == nil {
		for _, b := range []byte(o.S) {
			t.Putchar(b)
		}
	} else {
		t.LastPrinted = o.X.Eval(t)
		bb := []byte(strings.Trim(fmt.Sprintf("%.15g", t.LastPrinted), " ") + " ")
		for _, b := range bb {
			t.Putchar(b)
		}
	}
	return 0
}

type CallCmd struct {
	Var  string
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

type DimCmd struct {
	Var  string
	Dims []int
}

func (o *DimCmd) Eval(t *Terp) int {
	return 0
}
func (o *DimCmd) String() string { return F("DIM %s (%v)", o.Var, o.Dims) }

type LetCmd struct {
	Dest *Expr
	X    *Expr
}

func (o *LetCmd) Eval(t *Terp) int {
	x := o.X.Eval(t) // Eval rhs first.
	if o.Dest.Var == nil {
		name := *o.Dest.A.Var
		thing, ok := t.Arrays[name]
		if !ok {
			panic(F("No such array: %q", name))
		}
		subs := o.Dest.Subs
		for len(subs) > 0 {
			//println(F("thing = %v", thing))
			//println(F("sub[0]=%v", subs[0]))
			e := SnapToInt(subs[0].Eval(t))
			//println(F("e=%v", e))
			if e < 0 {
				panic(F("negative subscript: array %s got %d", name, e))
			}
			switch t := thing.(type) {
			case []interface{}:
				if len(subs) <= 1 {
					panic("too few subscripts for array: " + name)
				}
				if e > len(t) {
					panic(F("subscript too big: array %q got %d want <=%d", name, e, len(t)))
				}
				thing = t[e]
			case []float64:
				if len(subs) != 1 {
					panic("too many subscripts for array: " + name)
				}
				if e > len(t) {
					panic(F("subscript too big: array %q got %d want <=%d", name, e, len(t)))
				}
				t[e] = x
				return 0
			default:
			}
			subs = subs[1:]
		}
		//println(F("THING = %v", thing))
		panic(F("notreached: Subs=%v", o.Dest.Subs))
	} else {
		t.Scalars[*o.Dest.Var] = x
	}
	return 0
}
func (o *LetCmd) String() string { return F("LET %s = %v", o.Dest, o.X) }

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
	t.Scalars[o.Var] = fr.Value
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
	t.Scalars[o.Var] = begin
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
	if SnapToInt(cond) == 0 {
		return o.Else
	}
	return o.Then
}

type Cmd interface {
	Eval(t *Terp) int
	String() string
}

func (lex *Terp) ParsePrim() *Expr {
	switch lex.K {
	case Var:
		z := &Expr{}
		s := lex.S
		lex.Advance()
		z.Var = &s
		return z
	case Number:
		z := &Expr{}
		f := lex.F
		lex.Advance()
		z.Const = &f
		return z
	case Punc:
		if lex.S == "(" {
			lex.Advance()
			z := lex.ParseExpr()
			lex.ParseMustSym(")")
			return z
		}
	}
	panic("expected prim")
}
func (lex *Terp) ParseExpr() *Expr {
	return lex.ParseRelop()
}

var MatchProduct = regexp.MustCompile(`^([*]|/|%)$`).MatchString
var MatchSum = regexp.MustCompile(`^([+]|-|\^)$`).MatchString
var MatchRelop = regexp.MustCompile(`^(=|==|<|<=|>|>=|!=|<>)$`).MatchString

// ParseComposite handles subscripting, and someday function calls.
func (lex *Terp) ParseComposite() *Expr {
	a := lex.ParsePrim()
	if lex.S == "(" {
		lex.Advance() // Consume paren.
		var subs []*Expr
		for lex.S != ")" {

			b := lex.ParseExpr()
			subs = append(subs, b)

			if lex.S != "," && lex.S != ")" {
				panic("Expected , or ) after expr in subscript")
			}
			if lex.S == "," {
				lex.Advance()
			}
		}
		lex.Advance()
		a = &Expr{A: a, Subs: subs}
	}
	return a
}

func (lex *Terp) ParseProduct() *Expr {
	a := lex.ParseComposite()
	for lex.K == Op && MatchProduct(lex.S) {
		op := lex.S
		lex.Advance()             // Consume op.
		b := lex.ParseComposite() // May be the negative constant.
		a = &Expr{A: a, Op: op, B: b}
	}
	return a
}
func (lex *Terp) ParseSum() *Expr {
	a := lex.ParseProduct()
	for lex.K == Op && MatchSum(lex.S) || lex.K == Number && lex.S[0] == '-' {
		op := lex.S
		if lex.K == Number { // Negative constant (ambiguous "-" sign)
			op = "+"
		} else {
			lex.Advance() // Consume op.
		}
		b := lex.ParseProduct() // May be the negative constant.
		a = &Expr{A: a, Op: op, B: b}
	}
	return a
}
func (lex *Terp) ParseRelop() *Expr {
	a := lex.ParseSum()
	for lex.K == Op && MatchRelop(lex.S) {
		op := lex.S
		lex.Advance()       // Consume op.
		b := lex.ParseSum() // May be the negative constant.
		a = &Expr{A: a, Op: op, B: b}
	}
	return a
}

func (lex *Terp) ParseProgram() []Line {
	var n int // the line number
	defer func() {
		r := recover()
		if r != nil {
			r = fmt.Sprintf("Parse Error: %v:\n... Line %d, Before code: %q", r, n, lex.Program[lex.P:])
			fmt.Fprintf(os.Stderr, "\n$v\n", r)
			debug.PrintStack()
			panic(r)
		}
		return
	}()
	var c Cmd
	var w string // the command
	var z []Line
Loop:
	for {
		n++ // Increment default line number.
		for lex.K == Newline {
			lex.Advance()
		}
		// Want line number.
		switch lex.K {
		// default: use the default line number.
		case EOF:
			break Loop
		case Newline:
			continue
		case Number:
			n = int(lex.F)
			lex.Advance()
		}
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
		switch w {
		case ";":
			c = &NopCmd{}
		case "REM":
			c = &NopCmd{}
			for lex.K != EOF && lex.K != Newline {
				lex.Advance()
			}
		case "GOTO":
			n := int(lex.ParseNumber())
			c = &GotoCmd{N: n}
		case "GOSUB":
			n := int(lex.ParseNumber())
			c = &GosubCmd{CallLine: n}
		case "RETURN":
			c = &ReturnCmd{}
		case "PRINT":
			if lex.K == String {
				c = &PrintCmd{S: lex.S}
				lex.Advance()
			} else {
				x := lex.ParseExpr()
				c = &PrintCmd{X: x}
			}
		case "DIM":
			v := lex.ParseVar()
			lex.ParseMustSym("(")
			var dims []int
			for lex.S != ")" {
				n := int(lex.ParseNumber())
				dims = append(dims, n)
				if lex.S != ")" && lex.S != "," {
					panic("expected , or ) after dimension")
				}
				if lex.S == "," {
					lex.Advance()
				}
			}
			lex.Advance()
			c = &DimCmd{Var: v, Dims: dims}
		case "LET":
			v := lex.ParseDestination()
			lex.ParseMustSym("=")
			x := lex.ParseExpr()
			c = &LetCmd{Dest: v, X: x}
		case "NEXT":
			v := lex.ParseVar()
			c = &NextCmd{Var: v}
		case "STOP":
			c = &StopCmd{}
		case "END":
			c = &EndCmd{}
		case "FOR":
			v := lex.ParseVar()
			lex.ParseMustSym("=")
			x := lex.ParseExpr()
			lex.ParseMustKeyword("TO")
			y := lex.ParseExpr()
			c = &ForCmd{Var: v, Begin: x, End: y}
		case "IF":
			cond := lex.ParseExpr()
			lex.ParseMustKeyword("THEN")
			ifTrue := SnapToInt(lex.ParseNumber())
			ifFalse := 0
			if lex.S == "ELSE" {
				lex.ParseMustKeyword("ELSE")
				ifFalse = SnapToInt(lex.ParseNumber())
			}
			c = &IfCmd{Cond: cond, Then: ifTrue, Else: ifFalse}
		case "CALL":
			name := strings.ToLower(lex.ParseVar())
			lex.ParseMustSym("(")
			var args []*Expr
			for {
				for lex.S == "," {
					lex.Advance()
				}
				if lex.S == ")" {
					break
				}
				a := lex.ParseExpr()
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

func (lex *Terp) ParseDestination() *Expr {
	a := lex.ParseComposite()
	if a.Var != nil && *a.Var != "" {
		return a // Scalar destination.
	}
	if a.Subs != nil && a.A.Var != nil && *a.A.Var != "" {
		return a // Subscripted destination.
	}
	panic(F("Illegal destination in assignment: %v", a))
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
	Check((lex.K == Op || lex.K == Punc) && lex.S == x, "expected symbol: "+x)
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
