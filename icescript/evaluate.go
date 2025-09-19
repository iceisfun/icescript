package icescript

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ---------- Values & Host Interop ----------

type ValueKind int

const (
	NullKind ValueKind = iota
	IntKind
	FloatKind
	BoolKind
	StringKind
)

type Value struct {
	Kind ValueKind
	I    int64
	F    float64
	B    bool
	S    string
}

func VNull() Value           { return Value{Kind: NullKind} }
func VInt(v int64) Value     { return Value{Kind: IntKind, I: v} }
func VFloat(v float64) Value { return Value{Kind: FloatKind, F: v} }
func VBool(v bool) Value     { return Value{Kind: BoolKind, B: v} }
func VString(v string) Value { return Value{Kind: StringKind, S: v} }
func (v Value) String() string {
	switch v.Kind {
	case IntKind:
		return fmt.Sprintf("%d", v.I)
	case FloatKind:
		return fmt.Sprintf("%g", v.F)
	case BoolKind:
		if v.B {
			return "true"
		}
		return "false"
	case StringKind:
		return v.S
	default:
		return "null"
	}
}
func (v Value) AsFloat() float64 {
	switch v.Kind {
	case FloatKind:
		return v.F
	case IntKind:
		return float64(v.I)
	case BoolKind:
		if v.B {
			return 1
		}
		return 0
	case StringKind:
		f, _ := strconv.ParseFloat(v.S, 64)
		return f
	default:
		return 0
	}
}
func (v Value) AsInt() int64 {
	switch v.Kind {
	case IntKind:
		return v.I
	case FloatKind:
		return int64(v.F)
	case BoolKind:
		if v.B {
			return 1
		}
		return 0
	case StringKind:
		i, _ := strconv.ParseInt(v.S, 10, 64)
		return i
	default:
		return 0
	}
}
func (v Value) AsBool() bool {
	switch v.Kind {
	case BoolKind:
		return v.B
	case IntKind:
		return v.I != 0
	case FloatKind:
		return v.F != 0
	case StringKind:
		return v.S != ""
	default:
		return false
	}
}

type HostFunc func(*VM, []Value) (Value, error)

// ---------- AST ----------

type Node interface{ Pos() Position }

type Program struct {
	Funcs map[string]*FuncDecl
}

type FuncDecl struct {
	Name    string
	Params  []Param
	RetType string // parsed but not enforced (yet)
	Body    *BlockStmt
	Start   Position
}

func (f *FuncDecl) Pos() Position { return f.Start }

type Param struct {
	Name string
	Type string
	Posn Position
}

type Stmt interface {
	Node
	stmtNode()
}

type BlockStmt struct {
	Stmts []Stmt
	LPos  Position
	RPos  Position
}

func (*BlockStmt) stmtNode()       {}
func (b *BlockStmt) Pos() Position { return b.LPos }

type ReturnStmt struct {
	Expr Expr
	P    Position
}

func (*ReturnStmt) stmtNode()       {}
func (r *ReturnStmt) Pos() Position { return r.P }

type ExprStmt struct {
	X Expr
	P Position
}

func (*ExprStmt) stmtNode()       {}
func (e *ExprStmt) Pos() Position { return e.P }

type Expr interface {
	Node
	exprNode()
}

type Ident struct {
	Name string
	P    Position
}

func (*Ident) exprNode()       {}
func (i *Ident) Pos() Position { return i.P }

type NumberLit struct {
	Raw string
	P   Position
}

func (*NumberLit) exprNode()       {}
func (n *NumberLit) Pos() Position { return n.P }

type StringLit struct {
	Val string
	P   Position
}

func (*StringLit) exprNode()       {}
func (s *StringLit) Pos() Position { return s.P }

type BoolLit struct {
	Val bool
	P   Position
}

func (*BoolLit) exprNode()       {}
func (b *BoolLit) Pos() Position { return b.P }

type NullLit struct {
	P Position
}

func (*NullLit) exprNode()       {}
func (n *NullLit) Pos() Position { return n.P }

type CallExpr struct {
	Callee string
	Args   []Expr
	P      Position
}

func (*CallExpr) exprNode()       {}
func (c *CallExpr) Pos() Position { return c.P }

type BinaryExpr struct {
	Op    string
	Left  Expr
	Right Expr
	P     Position
}

func (*BinaryExpr) exprNode()       {}
func (b *BinaryExpr) Pos() Position { return b.P }

// ---------- Parser (Pratt for expressions) ----------

type Parser struct {
	lx      *Lexer
	cur     Token
	peekTok Token
	errs    []error
}

func NewParser(lx *Lexer) *Parser {
	p := &Parser{lx: lx}
	p.next()
	p.next()
	return p
}

func (p *Parser) next() { p.cur, p.peekTok = p.peekTok, p.lx.NextToken() }
func (p *Parser) expectLit(l string) bool {
	if p.cur.Literal == l {
		return true
	}
	p.errf(p.cur.Position, "expected %q, got %q", l, p.cur.Literal)
	return false
}
func (p *Parser) expectKind(k TokenKind) bool {
	if p.cur.Kind == k {
		return true
	}
	p.errf(p.cur.Position, "expected token kind %d, got %d (%q)", k, p.cur.Kind, p.cur.Literal)
	return false
}
func (p *Parser) errf(pos Position, f string, a ...any) {
	p.errs = append(p.errs, fmt.Errorf("%s:%d:%d: %s" /* filename? */, "script", pos.Line, pos.Column, fmt.Sprintf(f, a...)))
}

func (p *Parser) ParseProgram() (*Program, []error) {
	pr := &Program{Funcs: make(map[string]*FuncDecl)}
	for {
		// EOF?
		if p.cur.Kind == EOF || (p.cur.Kind == ILLEGAL && p.cur.Literal == "") {
			break
		}
		// skip stray tokens until "func"
		if p.cur.Literal != "func" && p.cur.Kind != FUNC {
			p.next()
			continue
		}
		fn := p.parseFunc()
		if fn != nil {
			pr.Funcs[fn.Name] = fn
		}
	}
	if len(p.errs) > 0 {
		return nil, p.errs
	}
	return pr, nil
}

func (p *Parser) parseFunc() *FuncDecl {
	start := p.cur.Position
	// consume "func"
	p.next()

	// name
	if !(p.cur.Kind == IDENT) {
		p.errf(p.cur.Position, "expected function name, got %q", p.cur.Literal)
		return nil
	}
	name := p.cur.Literal
	p.next()

	// params
	if p.cur.Literal != "(" {
		p.errf(p.cur.Position, "expected '(' after func name")
		return nil
	}
	p.next()

	var params []Param
	for p.cur.Literal != ")" && p.cur.Kind != EOF {
		// name
		if p.cur.Kind != IDENT {
			p.errf(p.cur.Position, "expected param name")
			return nil
		}
		pname := p.cur.Literal
		ppos := p.cur.Position
		p.next()
		// type (IDENT) — optional in practice; we accept IDENT if present
		ptype := ""
		if p.cur.Kind == IDENT {
			ptype = p.cur.Literal
			p.next()
		}
		params = append(params, Param{Name: pname, Type: ptype, Posn: ppos})
		// comma?
		if p.cur.Literal == "," {
			p.next()
		}
	}
	if p.cur.Literal != ")" {
		p.errf(p.cur.Position, "expected ')'")
		return nil
	}
	p.next()

	// optional return type (IDENT)
	retType := ""
	if p.cur.Kind == IDENT {
		retType = p.cur.Literal
		p.next()
	}

	// body
	if p.cur.Literal != "{" {
		p.errf(p.cur.Position, "expected '{' to start function body")
		return nil
	}
	body := p.parseBlock()

	return &FuncDecl{Name: name, Params: params, RetType: retType, Body: body, Start: start}
}

func (p *Parser) parseBlock() *BlockStmt {
	lpos := p.cur.Position
	p.next()
	var stmts []Stmt
	for p.cur.Literal != "}" && p.cur.Kind != EOF {
		st := p.parseStmt()
		if st != nil {
			stmts = append(stmts, st)
		} else {
			// resync on '}' or 'return'
			if p.cur.Literal != "}" {
				p.next()
			}
		}
	}
	rpos := p.cur.Position
	if p.cur.Literal == "}" {
		p.next()
	}
	return &BlockStmt{Stmts: stmts, LPos: lpos, RPos: rpos}
}

func (p *Parser) parseStmt() Stmt {
	// return
	if p.cur.Literal == "return" || p.cur.Kind == RETURN {
		pos := p.cur.Position
		p.next()
		ex := p.parseExpr(0)
		// optional semicolon
		if p.cur.Literal == ";" {
			p.next()
		}
		return &ReturnStmt{Expr: ex, P: pos}
	}
	// expression statement
	pos := p.cur.Position
	ex := p.parseExpr(0)
	// optional semicolon
	if p.cur.Literal == ";" {
		p.next()
	}
	return &ExprStmt{X: ex, P: pos}
}

// precedence table
func prec(op string) int {
	switch op {
	case "||":
		return 1
	case "&&":
		return 2
	case "==", "!=", "<", ">", "<=", ">=":
		return 3
	case "+", "-":
		return 4
	case "*", "/", "%":
		return 5
	default:
		return 0
	}
}

func (p *Parser) parseExpr(minPrec int) Expr {
	left := p.parsePrimary()
	for {
		op := p.cur.Literal
		pk := p.cur.Kind
		// operators generally come as literal symbols; some lexers set specific kinds for EQ/NEQ etc.
		if !isBinaryOpLiteral(op) && !isBinaryOpKind(pk) {
			break
		}
		// normalize op for kind-based tokens
		if op == "" {
			op = kindToOp(pk)
		}
		pPre := prec(op)
		if pPre < minPrec {
			break
		}
		// consume op
		opTok := p.cur
		p.next()

		right := p.parseExpr(pPre + 1)
		left = &BinaryExpr{Op: op, Left: left, Right: right, P: opTok.Position}
	}
	return left
}

func isBinaryOpLiteral(s string) bool {
	switch s {
	case "+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||":
		return true
	default:
		return false
	}
}

func isBinaryOpKind(k TokenKind) bool {
	switch k {
	case PLUS, MINUS, STAR, SLASH, PERCENT, EQ, NEQ, LT, GT, LTE, GTE, AND, OR:
		return true
	default:
		return false
	}
}

func kindToOp(k TokenKind) string {
	switch k {
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case STAR:
		return "*"
	case SLASH:
		return "/"
	case PERCENT:
		return "%"
	case EQ:
		return "=="
	case NEQ:
		return "!="
	case LT:
		return "<"
	case GT:
		return ">"
	case LTE:
		return "<="
	case GTE:
		return ">="
	case AND:
		return "&&"
	case OR:
		return "||"
	default:
		return ""
	}
}

func (p *Parser) parsePrimary() Expr {
	t := p.cur
	switch {
	case t.Kind == IDENT:
		// could be ident or call
		name := t.Literal
		p.next()
		if p.cur.Literal == "(" {
			callPos := t.Position
			p.next()
			var args []Expr
			for p.cur.Literal != ")" && p.cur.Kind != EOF {
				args = append(args, p.parseExpr(0))
				if p.cur.Literal == "," {
					p.next()
				}
			}
			if p.cur.Literal != ")" {
				p.errf(p.cur.Position, "expected ')'")
			} else {
				p.next()
			}
			return &CallExpr{Callee: name, Args: args, P: callPos}
		}
		return &Ident{Name: name, P: t.Position}

	case t.Kind == NUMBER:
		p.next()
		return &NumberLit{Raw: t.Literal, P: t.Position}

	case t.Kind == STRING:
		p.next()
		return &StringLit{Val: t.Literal, P: t.Position}

	case t.Kind == TRUE || strings.EqualFold(t.Literal, "true"):
		p.next()
		return &BoolLit{Val: true, P: t.Position}

	case t.Kind == FALSE || strings.EqualFold(t.Literal, "false"):
		p.next()
		return &BoolLit{Val: false, P: t.Position}

	case t.Kind == NULL || strings.EqualFold(t.Literal, "null"):
		p.next()
		return &NullLit{P: t.Position}

	case t.Literal == "(":
		p.next()
		ex := p.parseExpr(0)
		if p.cur.Literal != ")" {
			p.errs = append(p.errs, fmt.Errorf("script:%d:%d: expected ')'", p.cur.Position.Line, p.cur.Position.Column))
		} else {
			p.next()
		}
		return ex
	default:
		p.errf(t.Position, "unexpected token %q", t.Literal)
		p.next()
		return &NullLit{P: t.Position}
	}
}

// ---------- VM / Evaluation ----------

type VM struct {
	prog      *Program
	globals   map[string]Value
	hostFuncs map[string]HostFunc
	callstack []callFrame
}

type callFrame struct {
	funcName string
	pos      Position
}

func NewVM() *VM {
	return &VM{
		globals:   make(map[string]Value),
		hostFuncs: make(map[string]HostFunc),
	}
}

func (vm *VM) RegisterHostFunc(name string, fn HostFunc) { vm.hostFuncs[name] = fn }
func (vm *VM) SetGlobal(name string, v Value)            { vm.globals[name] = v }

func (vm *VM) LoadProgram(p *Program) { vm.prog = p }

func (vm *VM) Invoke(name string, args ...Value) (Value, error) {
	// script function?
	if vm.prog != nil {
		if f, ok := vm.prog.Funcs[name]; ok {
			if len(args) != len(f.Params) {
				return VNull(), vm.rtErr(f.Start, "function %s expects %d args, got %d", f.Name, len(f.Params), len(args))
			}
			env := make(map[string]Value)
			for i, prm := range f.Params {
				env[prm.Name] = args[i]
			}
			vm.callstack = append(vm.callstack, callFrame{funcName: f.Name, pos: f.Start})
			val, err := vm.execBlock(env, f.Body)
			vm.callstack = vm.callstack[:len(vm.callstack)-1]
			return val, err
		}
	}
	// host function?
	if hf, ok := vm.hostFuncs[name]; ok {
		return hf(vm, args)
	}
	return VNull(), vm.rtErr(Position{}, "unknown function %q", name)
}

func (vm *VM) execBlock(env map[string]Value, blk *BlockStmt) (Value, error) {
	var last Value = VNull()
	for _, st := range blk.Stmts {
		switch s := st.(type) {
		case *ReturnStmt:
			val, err := vm.evalExpr(env, s.Expr)
			if err != nil {
				return VNull(), err
			}
			return val, nil
		case *ExprStmt:
			v, err := vm.evalExpr(env, s.X)
			if err != nil {
				return VNull(), err
			}
			last = v
		default:
			return VNull(), vm.rtErr(st.Pos(), "unsupported statement")
		}
	}
	return last, nil
}

func (vm *VM) evalExpr(env map[string]Value, e Expr) (Value, error) {
	switch ex := e.(type) {
	case *Ident:
		// locals > globals
		if v, ok := env[ex.Name]; ok {
			return v, nil
		}
		if v, ok := vm.globals[ex.Name]; ok {
			return v, nil
		}
		return VNull(), vm.rtErr(ex.P, "undefined identifier %q", ex.Name)

	case *NumberLit:
		if strings.ContainsAny(ex.Raw, ".eE") {
			f, err := strconv.ParseFloat(ex.Raw, 64)
			if err != nil {
				return VNull(), vm.rtErr(ex.P, "invalid float literal %q", ex.Raw)
			}
			return VFloat(f), nil
		}
		i, err := strconv.ParseInt(ex.Raw, 10, 64)
		if err != nil {
			return VNull(), vm.rtErr(ex.P, "invalid int literal %q", ex.Raw)
		}
		return VInt(i), nil

	case *StringLit:
		return VString(ex.Val), nil

	case *BoolLit:
		return VBool(ex.Val), nil

	case *NullLit:
		return VNull(), nil

	case *CallExpr:
		// scripted function?
		if vm.prog != nil {
			if f, ok := vm.prog.Funcs[ex.Callee]; ok {
				if len(ex.Args) != len(f.Params) {
					return VNull(), vm.rtErr(ex.P, "function %s expects %d args, got %d", f.Name, len(f.Params), len(ex.Args))
				}
				argv := make([]Value, len(ex.Args))
				for i, a := range ex.Args {
					v, err := vm.evalExpr(env, a)
					if err != nil {
						return VNull(), err
					}
					argv[i] = v
				}
				vm.callstack = append(vm.callstack, callFrame{funcName: f.Name, pos: f.Start})
				childEnv := make(map[string]Value, len(f.Params))
				for i, prm := range f.Params {
					childEnv[prm.Name] = argv[i]
				}
				val, err := vm.execBlock(childEnv, f.Body)
				vm.callstack = vm.callstack[:len(vm.callstack)-1]
				return val, err
			}
		}
		// host function?
		if hf, ok := vm.hostFuncs[ex.Callee]; ok {
			argv := make([]Value, len(ex.Args))
			for i, a := range ex.Args {
				v, err := vm.evalExpr(env, a)
				if err != nil {
					return VNull(), err
				}
				argv[i] = v
			}
			return hf(vm, argv)
		}
		return VNull(), vm.rtErr(ex.P, "unknown function %q", ex.Callee)

	case *BinaryExpr:
		lv, err := vm.evalExpr(env, ex.Left)
		if err != nil {
			return VNull(), err
		}
		rv, err := vm.evalExpr(env, ex.Right)
		if err != nil {
			return VNull(), err
		}

		switch ex.Op {
		case "+":
			// string concat if either side is string
			if lv.Kind == StringKind || rv.Kind == StringKind {
				return VString(lv.String() + rv.String()), nil
			}
			// numeric add (preserve float if any side is float)
			if lv.Kind == FloatKind || rv.Kind == FloatKind {
				return VFloat(lv.AsFloat() + rv.AsFloat()), nil
			}
			return VInt(lv.AsInt() + rv.AsInt()), nil
		case "-":
			if lv.Kind == FloatKind || rv.Kind == FloatKind {
				return VFloat(lv.AsFloat() - rv.AsFloat()), nil
			}
			return VInt(lv.AsInt() - rv.AsInt()), nil
		case "*":
			if lv.Kind == FloatKind || rv.Kind == FloatKind {
				return VFloat(lv.AsFloat() * rv.AsFloat()), nil
			}
			return VInt(lv.AsInt() * rv.AsInt()), nil
		case "/":
			lf := lv.AsFloat()
			rf := rv.AsFloat()
			if rf == 0 {
				return VNull(), vm.rtErr(ex.P, "division by zero")
			}
			// favor float division
			return VFloat(lf / rf), nil
		case "%":
			ri := rv.AsInt()
			if ri == 0 {
				return VNull(), vm.rtErr(ex.P, "modulo by zero")
			}
			return VInt(lv.AsInt() % ri), nil
		case "==":
			return VBool(lv.String() == rv.String()), nil
		case "!=":
			return VBool(lv.String() != rv.String()), nil
		case "<":
			return VBool(lv.AsFloat() < rv.AsFloat()), nil
		case "<=":
			return VBool(lv.AsFloat() <= rv.AsFloat()), nil
		case ">":
			return VBool(lv.AsFloat() > rv.AsFloat()), nil
		case ">=":
			return VBool(lv.AsFloat() >= rv.AsFloat()), nil
		case "&&":
			return VBool(lv.AsBool() && rv.AsBool()), nil
		case "||":
			return VBool(lv.AsBool() || rv.AsBool()), nil
		default:
			return VNull(), vm.rtErr(ex.P, "unsupported operator %q", ex.Op)
		}
	default:
		return VNull(), vm.rtErr(e.Pos(), "unsupported expression")
	}
}

func (vm *VM) rtErr(pos Position, f string, a ...any) error {
	msg := fmt.Sprintf(f, a...)
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "runtime error at %d:%d: %s", pos.Line, pos.Column, msg)
	// stack (most recent last)
	for i := len(vm.callstack) - 1; i >= 0; i-- {
		fr := vm.callstack[i]
		fmt.Fprintf(sb, "\n  at %s (%d:%d)", fr.funcName, fr.pos.Line, fr.pos.Column)
	}
	return errors.New(sb.String())
}

// ---------- Convenience: Parse + Load + Invoke ----------

// Compile parses source into a Program and loads it into the VM.
func (vm *VM) Compile(src string) error {
	p := NewParser(New(src))
	pr, errs := p.ParseProgram()
	if len(errs) > 0 {
		var b strings.Builder
		for _, e := range errs {
			b.WriteString(e.Error())
			b.WriteByte('\n')
		}
		return errors.New(strings.TrimSpace(b.String()))
	}
	vm.LoadProgram(pr)
	return nil
}
