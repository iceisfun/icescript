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
	ArrayKind  // NEW
	ObjectKind // NEW
)

type Value struct {
	Kind      ValueKind
	I         int64
	F         float64
	B         bool
	S         string
	Arr       []Value
	Obj       map[string]Value
	ConstName string
}

func VArray(xs []Value) Value          { return Value{Kind: ArrayKind, Arr: xs} }
func VObject(m map[string]Value) Value { return Value{Kind: ObjectKind, Obj: m} }
func VNull() Value                     { return Value{Kind: NullKind} }
func VInt(v int64) Value               { return Value{Kind: IntKind, I: v} }
func VFloat(v float64) Value           { return Value{Kind: FloatKind, F: v} }
func VBool(v bool) Value               { return Value{Kind: BoolKind, B: v} }
func VString(v string) Value           { return Value{Kind: StringKind, S: v} }
func (v Value) String() string {
	var base string
	switch v.Kind {
	case IntKind:
		base = fmt.Sprintf("%d", v.I)
	case FloatKind:
		base = fmt.Sprintf("%g", v.F)
	case BoolKind:
		if v.B {
			base = "true"
		} else {
			base = "false"
		}
	case StringKind:
		base = v.S
	case ArrayKind:
		var b strings.Builder
		b.WriteByte('[')
		for i, e := range v.Arr {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(e.String())
		}
		b.WriteByte(']')
		base = b.String()
	case ObjectKind:
		var b strings.Builder
		b.WriteByte('{')
		i := 0
		for k, val := range v.Obj {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(k)
			b.WriteByte(':')
			b.WriteString(val.String())
			i++
		}
		b.WriteByte('}')
		base = b.String()
	default:
		base = "null"
	}
	if v.ConstName != "" {
		return base + "::" + v.ConstName
	}
	return base
}

type environment struct {
	vars   map[string]Value
	parent *environment
}

func newEnvironment(parent *environment) *environment {
	return &environment{vars: make(map[string]Value), parent: parent}
}

func (e *environment) declare(name string, v Value) {
	if e.vars == nil {
		e.vars = make(map[string]Value)
	}
	e.vars[name] = v
}

func (e *environment) lookup(name string) (Value, bool) {
	for cur := e; cur != nil; cur = cur.parent {
		if cur.vars != nil {
			if v, ok := cur.vars[name]; ok {
				return v, true
			}
		}
	}
	return Value{}, false
}

func (e *environment) assign(name string, v Value) bool {
	for cur := e; cur != nil; cur = cur.parent {
		if cur.vars != nil {
			if _, ok := cur.vars[name]; ok {
				cur.vars[name] = v
				return true
			}
		}
	}
	return false
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
	Name   string
	Params []Param
	Body   *BlockStmt
	Start  Position
}

func (f *FuncDecl) Pos() Position { return f.Start }

type Param struct {
	Name string
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

type BreakStmt struct {
	P Position
}

func (*BreakStmt) stmtNode()       {}
func (b *BreakStmt) Pos() Position { return b.P }

type ContinueStmt struct {
	P Position
}

func (*ContinueStmt) stmtNode()       {}
func (c *ContinueStmt) Pos() Position { return c.P }

type loopSignal int

const (
	signalBreak loopSignal = iota
	signalContinue
)

type loopErr struct {
	kind loopSignal
	pos  Position
}

func (e loopErr) Error() string {
	switch e.kind {
	case signalBreak:
		return "break"
	case signalContinue:
		return "continue"
	default:
		return "loop control"
	}
}

type returnErr struct {
	value Value
}

func (e returnErr) Error() string { return "return" }

func convertLoopErr(vm *VM, err error) error {
	var le loopErr
	if errors.As(err, &le) {
		switch le.kind {
		case signalBreak:
			return vm.rtErr(le.pos, "break outside of loop")
		case signalContinue:
			return vm.rtErr(le.pos, "continue outside of loop")
		}
	}
	return err
}

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

type UnaryExpr struct {
	Op string
	X  Expr
	P  Position
}

func (*UnaryExpr) exprNode()       {}
func (u *UnaryExpr) Pos() Position { return u.P }

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
	lx                   *Lexer
	cur                  Token
	peekTok              Token
	errs                 []error
	noObjectLiteralDepth int
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
		if p.cur.Kind == EOF {
			break
		}
		if p.cur.Kind == ILLEGAL {
			p.errf(p.cur.Position, "unexpected token %q", p.cur.Literal)
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
		if p.cur.Kind != IDENT {
			p.errf(p.cur.Position, "expected parameter name, got %q", p.cur.Literal)
			return nil
		}
		pname := p.cur.Literal
		ppos := p.cur.Position
		p.next()
		params = append(params, Param{Name: pname, Posn: ppos})
		switch p.cur.Literal {
		case ",":
			p.next()
		case ")":
			// loop will exit
		default:
			p.errf(p.cur.Position, "unexpected token %q after parameter %s; types are not supported", p.cur.Literal, pname)
			return nil
		}
	}
	if p.cur.Literal != ")" {
		p.errf(p.cur.Position, "expected ')'")
		return nil
	}
	p.next()

	if !(p.cur.Literal == "{" || p.cur.Kind == LBRACE) {
		if p.cur.Kind == IDENT {
			p.errf(p.cur.Position, "unexpected token %q before function body; return types are not supported", p.cur.Literal)
			return nil
		}
		p.errf(p.cur.Position, "expected '{' to start function body")
		return nil
	}
	body := p.parseBlock()

	return &FuncDecl{Name: name, Params: params, Body: body, Start: start}
}

func (p *Parser) parseBlock() *BlockStmt {
	lpos := p.cur.Position
	p.next()
	var stmts []Stmt
	for {
		switch {
		case p.cur.Literal == "}":
			rpos := p.cur.Position
			p.next()
			return &BlockStmt{Stmts: stmts, LPos: lpos, RPos: rpos}
		case p.cur.Kind == EOF:
			p.errf(p.cur.Position, "expected '}' to close block")
			return &BlockStmt{Stmts: stmts, LPos: lpos, RPos: p.cur.Position}
		default:
			st := p.parseStmt()
			if st != nil {
				stmts = append(stmts, st)
			} else if p.cur.Kind != EOF {
				p.next()
			}
		}
	}

}

func (p *Parser) parseStmt() Stmt {
	// return
	if p.cur.Literal == "return" || p.cur.Kind == RETURN {
		pos := p.cur.Position
		p.next()
		ex := p.parseExpr(0)
		if p.cur.Literal == ";" {
			p.next()
		}
		return &ReturnStmt{Expr: ex, P: pos}
	}

	if p.cur.Kind == BREAK {
		pos := p.cur.Position
		p.next()
		if p.cur.Literal == ";" {
			p.next()
		}
		return &BreakStmt{P: pos}
	}

	if p.cur.Kind == CONTINUE {
		pos := p.cur.Position
		p.next()
		if p.cur.Literal == ";" {
			p.next()
		}
		return &ContinueStmt{P: pos}
	}

	// var name (= expr)?
	if p.cur.Literal == "var" || p.cur.Kind == VAR {
		pos := p.cur.Position
		p.next()
		if p.cur.Kind != IDENT {
			p.errf(p.cur.Position, "expected identifier after var")
			return nil
		}
		name := p.cur.Literal
		p.next()
		var init Expr = &NullLit{P: pos}
		if p.cur.Literal == "=" {
			p.next()
			init = p.parseExpr(0)
		}
		if p.cur.Literal == ";" {
			p.next()
		}
		return &VarStmt{Name: name, Init: init, P: pos}
	}

	// if condition { ... } (optional else)
	if p.cur.Literal == "if" || p.cur.Kind == IF {
		pos := p.cur.Position
		p.next()
		parened := false
		if p.cur.Literal == "(" {
			parened = true
			p.next()
		}
		p.noObjectLiteralDepth++
		cond := p.parseExpr(0)
		p.noObjectLiteralDepth--
		if parened {
			if !(p.cur.Literal == ")" || p.cur.Kind == RPAREN) {
				p.errf(p.cur.Position, "expected ')' to close if condition")
				return nil
			}
			p.next()
		}
		if !(p.cur.Literal == "{" || p.cur.Kind == LBRACE) {
			p.errf(p.cur.Position, "expected '{' to start if body")
			return nil
		}
		thenBlk := p.parseBlock()
		var elseBlk *BlockStmt
		if p.cur.Literal == "else" || p.cur.Kind == ELSE {
			p.next()
			if !(p.cur.Literal == "{" || p.cur.Kind == LBRACE) {
				p.errf(p.cur.Position, "expected '{' to start else body")
				return &IfStmt{Cond: cond, Then: thenBlk, P: pos}
			}
			elseBlk = p.parseBlock()
		}
		return &IfStmt{Cond: cond, Then: thenBlk, Else: elseBlk, P: pos}
	}

	// for x in iterable { ... }
	if p.cur.Literal == "for" || p.cur.Kind == FOR {
		pos := p.cur.Position
		p.next()
		if p.cur.Kind != IDENT {
			p.errf(p.cur.Position, "expected loop variable name")
			return nil
		}
		varName := p.cur.Literal
		p.next()
		if p.cur.Kind != IN {
			p.errf(p.cur.Position, "expected 'in' after loop variable")
			return nil
		}
		p.next()
		iter := p.parseExpr(0)
		if p.cur.Literal != "{" {
			p.errf(p.cur.Position, "expected '{' to start for-body")
			return nil
		}
		body := p.parseBlock()
		return &ForInStmt{VarName: varName, Iterable: iter, Body: body, P: pos}
	}

	// expression or assignment
	ex := p.parseExpr(0)
	if p.cur.Literal == "=" {
		assignable := isAssignableExpr(ex)
		pos := ex.Pos()
		eqTok := p.cur
		p.next()
		val := p.parseExpr(0)
		if p.cur.Literal == ";" {
			p.next()
		}
		if !assignable {
			p.errf(eqTok.Position, "invalid assignment target")
			return &ExprStmt{X: ex, P: pos}
		}
		return &AssignStmt{Target: ex, Value: val, P: pos}
	}
	if p.cur.Literal == ";" {
		p.next()
	}
	return &ExprStmt{X: ex, P: ex.Pos()}
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
	left := p.parseUnary()
	for {
		if p.noObjectLiteralDepth > 0 && (p.cur.Literal == "{" || p.cur.Kind == LBRACE) {
			break
		}
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

func (p *Parser) parseUnary() Expr {
	t := p.cur
	switch t.Kind {
	case PLUS, MINUS, BANG:
		op := t.Literal
		if op == "" {
			op = kindToOp(t.Kind)
		}
		p.next()
		x := p.parseUnary()
		return &UnaryExpr{Op: op, X: x, P: t.Position}
	default:
		return p.parsePrimary()
	}
}

func isAssignableExpr(e Expr) bool {
	switch e.(type) {
	case *Ident, *MemberExpr, *IndexExpr:
		return true
	default:
		return false
	}
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
	case BANG:
		return "!"
	case OR:
		return "||"
	default:
		return ""
	}
}

func (p *Parser) parsePrimary() Expr {
	var left Expr
	t := p.cur
	switch {
	case t.Kind == IDENT:
		name := t.Literal
		p.next()
		// function call ident(...)
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
			left = &CallExpr{Callee: name, Args: args, P: callPos}
		} else {
			left = &Ident{Name: name, P: t.Position}
		}

	case t.Kind == NUMBER:
		p.next()
		left = &NumberLit{Raw: t.Literal, P: t.Position}

	case t.Kind == STRING:
		p.next()
		left = &StringLit{Val: t.Literal, P: t.Position}

	case t.Kind == TRUE || strings.EqualFold(t.Literal, "true"):
		p.next()
		left = &BoolLit{Val: true, P: t.Position}

	case t.Kind == FALSE || strings.EqualFold(t.Literal, "false"):
		p.next()
		left = &BoolLit{Val: false, P: t.Position}

	case t.Kind == NULL || strings.EqualFold(t.Literal, "null"):
		p.next()
		left = &NullLit{P: t.Position}

	case t.Literal == "(":
		p.next()
		left = p.parseExpr(0)
		if p.cur.Literal != ")" {
			p.errf(p.cur.Position, "expected ')'")
		} else {
			p.next()
		}

	case t.Literal == "[":
		// array literal
		start := t.Position
		p.next()
		var elems []Expr
		for p.cur.Literal != "]" && p.cur.Kind != EOF {
			elems = append(elems, p.parseExpr(0))
			if p.cur.Literal == "," {
				p.next()
			}
		}
		if p.cur.Literal != "]" {
			p.errf(p.cur.Position, "expected ']'")
		} else {
			p.next()
		}
		left = &ArrayLit{Elems: elems, P: start}

	case t.Literal == "{":
		// object literal: { ident : expr, ... }
		start := t.Position
		p.next()
		var fields []Field
		for p.cur.Literal != "}" && p.cur.Kind != EOF {
			if p.cur.Kind != IDENT {
				p.errf(p.cur.Position, "expected field name")
				break
			}
			fname := p.cur.Literal
			fpos := p.cur.Position
			p.next()
			if p.cur.Literal != ":" {
				p.errf(p.cur.Position, "expected ':' after field name")
				break
			}
			p.next()
			fexpr := p.parseExpr(0)
			fields = append(fields, Field{Name: fname, Expr: fexpr, P: fpos})
			if p.cur.Literal == "," {
				p.next()
			}
		}
		if p.cur.Literal != "}" {
			p.errf(p.cur.Position, "expected '}'")
		} else {
			p.next()
		}
		left = &ObjectLit{Fields: fields, P: start}

	default:
		p.errf(t.Position, "unexpected token %q", t.Literal)
		p.next()
		left = &NullLit{P: t.Position}
	}

	// postfix: .field  and  [expr]
	for {
		switch p.cur.Literal {
		case ".":
			p.next()
			if p.cur.Kind != IDENT {
				p.errf(p.cur.Position, "expected identifier after '.'")
				return left
			}
			nameTok := p.cur
			p.next()
			left = &MemberExpr{Object: left, Name: nameTok.Literal, P: nameTok.Position}
		case "[":
			pos := p.cur.Position
			p.next()
			idx := p.parseExpr(0)
			if p.cur.Literal != "]" {
				p.errf(p.cur.Position, "expected ']'")
			} else {
				p.next()
			}
			left = &IndexExpr{Seq: left, Index: idx, P: pos}
		default:
			return left
		}
	}
}

// ---------- VM / Evaluation ----------

type VM struct {
	prog      *Program
	globals   map[string]Value
	constants map[string]Value
	hostFuncs map[string]HostFunc
	callstack []callFrame
}

type callFrame struct {
	funcName string
	pos      Position
}

func NewVM() *VM {
	vm := &VM{
		globals:   make(map[string]Value),
		hostFuncs: make(map[string]HostFunc),
	}
	vm.constants = make(map[string]Value)
	installBuiltins(vm)
	return vm
}

func (vm *VM) RegisterHostFunc(name string, fn HostFunc) { vm.hostFuncs[name] = fn }
func (vm *VM) SetGlobal(name string, v Value)            { vm.globals[name] = v }
func (vm *VM) SetConstants(consts map[string]any) error {
	vm.constants = make(map[string]Value, len(consts))
	for name, raw := range consts {
		var val Value
		switch c := raw.(type) {
		case Value:
			val = c
		default:
			converted, err := VFromGo(c)
			if err != nil {
				return fmt.Errorf("constant %s: %w", name, err)
			}
			val = converted
		}
		val.ConstName = name
		vm.constants[name] = val
	}
	return nil
}
func (vm *VM) GetGlobal(name string) Value {
	if v, ok := vm.globals[name]; ok {
		return v
	}
	return VNull()
}

func (vm *VM) LoadProgram(p *Program) { vm.prog = p }

func (vm *VM) Invoke(name string, args ...Value) (Value, error) {
	// script function?
	if vm.prog != nil {
		if f, ok := vm.prog.Funcs[name]; ok {
			if len(args) != len(f.Params) {
				return VNull(), vm.rtErr(f.Start, "function %s expects %d args, got %d", f.Name, len(f.Params), len(args))
			}
			env := newEnvironment(nil)
			for i, prm := range f.Params {
				env.declare(prm.Name, args[i])
			}
			vm.callstack = append(vm.callstack, callFrame{funcName: f.Name, pos: f.Start})
			val, err := vm.execBlock(env, f.Body)
			vm.callstack = vm.callstack[:len(vm.callstack)-1]
			if err != nil {
				var re returnErr
				if errors.As(err, &re) {
					return re.value, nil
				}
				return VNull(), convertLoopErr(vm, err)
			}
			return val, nil
		}
	}
	// host function?
	if hf, ok := vm.hostFuncs[name]; ok {
		return hf(vm, args)
	}
	return VNull(), vm.rtErr(Position{}, "unknown function %q", name)
}

func (vm *VM) execBlock(env *environment, blk *BlockStmt) (Value, error) {
	blockEnv := newEnvironment(env)
	var last Value = VNull()
	for _, st := range blk.Stmts {
		switch s := st.(type) {
		case *ReturnStmt:
			var val Value
			if s.Expr != nil {
				var err error
				val, err = vm.evalExpr(blockEnv, s.Expr)
				if err != nil {
					return VNull(), err
				}
			} else {
				val = VNull()
			}
			return val, returnErr{value: val}
		case *BreakStmt:
			return VNull(), loopErr{kind: signalBreak, pos: s.P}
		case *ContinueStmt:
			return VNull(), loopErr{kind: signalContinue, pos: s.P}

		case *ExprStmt:
			v, err := vm.evalExpr(blockEnv, s.X)
			if err != nil {
				return VNull(), err
			}
			last = v

		case *VarStmt:
			v, err := vm.evalExpr(blockEnv, s.Init)
			if err != nil {
				return VNull(), err
			}
			blockEnv.declare(s.Name, v)

		case *IfStmt:
			cond, err := vm.evalExpr(blockEnv, s.Cond)
			if err != nil {
				return VNull(), err
			}
			var branch *BlockStmt
			if cond.AsBool() {
				branch = s.Then
			} else {
				branch = s.Else
			}
			if branch != nil {
				val, err := vm.execBlock(blockEnv, branch)
				if err != nil {
					var re returnErr
					if errors.As(err, &re) {
						return val, err
					}
					return VNull(), err
				}
				last = val
			} else {
				last = VNull()
			}

		case *AssignStmt:
			v, err := vm.evalExpr(blockEnv, s.Value)
			if err != nil {
				return VNull(), err
			}
			if err := vm.assignValue(blockEnv, s.Target, v); err != nil {
				return VNull(), err
			}

		case *ForInStmt:
			iter, err := vm.evalExpr(blockEnv, s.Iterable)
			if err != nil {
				return VNull(), err
			}
			switch iter.Kind {
			case ArrayKind:
			loop:
				for _, item := range iter.Arr {
					if !blockEnv.assign(s.VarName, item) {
						blockEnv.declare(s.VarName, item)
					}
					val, err := vm.execBlock(blockEnv, s.Body)
					if err != nil {
						var re returnErr
						if errors.As(err, &re) {
							return val, err
						}
						var le loopErr
						if errors.As(err, &le) {
							switch le.kind {
							case signalContinue:
								continue
							case signalBreak:
								break loop
							}
						}
						return VNull(), err
					}
					last = val
				}
			default:
				return VNull(), vm.rtErr(s.P, "cannot iterate over %s", kindName(iter.Kind))
			}

		default:
			return VNull(), vm.rtErr(st.Pos(), "unsupported statement")
		}
	}
	return last, nil
}

func kindName(k ValueKind) string {
	switch k {
	case NullKind:
		return "null"
	case IntKind:
		return "int"
	case FloatKind:
		return "float"
	case BoolKind:
		return "bool"
	case StringKind:
		return "string"
	case ArrayKind:
		return "array"
	case ObjectKind:
		return "object"
	default:
		return "unknown"
	}
}

func (vm *VM) assignValue(env *environment, target Expr, value Value) error {
	switch t := target.(type) {
	case *Ident:
		if _, ok := vm.constants[t.Name]; ok {
			return vm.rtErr(t.P, "cannot assign to constant %q", t.Name)
		}
		if env != nil && env.assign(t.Name, value) {
			return nil
		}
		if _, ok := vm.globals[t.Name]; ok {
			vm.globals[t.Name] = value
			return nil
		}
		return vm.rtErr(t.P, "assignment to undefined identifier %q", t.Name)
	case *MemberExpr:
		obj, err := vm.evalExpr(env, t.Object)
		if err != nil {
			return err
		}
		if obj.ConstName != "" {
			return vm.rtErr(t.P, "cannot modify constant %q", obj.ConstName)
		}
		if obj.Kind != ObjectKind {
			return vm.rtErr(t.P, "cannot set field %q on %s", t.Name, kindName(obj.Kind))
		}
		if obj.Obj == nil {
			obj.Obj = make(map[string]Value)
		}
		obj.Obj[t.Name] = value
		switch parent := t.Object.(type) {
		case *Ident:
			if env != nil && env.assign(parent.Name, obj) {
				return nil
			}
			if _, ok := vm.globals[parent.Name]; ok {
				vm.globals[parent.Name] = obj
			}
		case *MemberExpr:
			return vm.assignValue(env, parent, obj)
		case *IndexExpr:
			return vm.assignValue(env, parent, obj)
		}
		return nil
	case *IndexExpr:
		seq, err := vm.evalExpr(env, t.Seq)
		if err != nil {
			return err
		}
		if seq.ConstName != "" {
			return vm.rtErr(t.P, "cannot modify constant %q", seq.ConstName)
		}
		idxVal, err := vm.evalExpr(env, t.Index)
		if err != nil {
			return err
		}
		idx := int(idxVal.AsInt())
		switch seq.Kind {
		case ArrayKind:
			if idx < 0 || idx >= len(seq.Arr) {
				return vm.rtErr(t.P, "index out of range")
			}
			seq.Arr[idx] = value
		default:
			return vm.rtErr(t.P, "cannot assign into %s with index", kindName(seq.Kind))
		}
		switch parent := t.Seq.(type) {
		case *Ident:
			if env != nil && env.assign(parent.Name, seq) {
				return nil
			}
			if _, ok := vm.globals[parent.Name]; ok {
				vm.globals[parent.Name] = seq
			}
		case *MemberExpr:
			return vm.assignValue(env, parent, seq)
		case *IndexExpr:
			return vm.assignValue(env, parent, seq)
		}
		return nil
	default:
		return vm.rtErr(target.Pos(), "invalid assignment target")
	}
}

func (vm *VM) evalExpr(env *environment, e Expr) (Value, error) {
	switch ex := e.(type) {
	case *Ident:
		// locals > globals
		if env != nil {
			if v, ok := env.lookup(ex.Name); ok {
				return v, nil
			}
		}
		if v, ok := vm.constants[ex.Name]; ok {
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

	case *UnaryExpr:
		val, err := vm.evalExpr(env, ex.X)
		if err != nil {
			return VNull(), err
		}
		switch ex.Op {
		case "-":
			if val.Kind == FloatKind {
				return VFloat(-val.AsFloat()), nil
			}
			return VInt(-val.AsInt()), nil
		case "+":
			return val, nil
		case "!":
			return VBool(!val.AsBool()), nil
		default:
			return VNull(), vm.rtErr(ex.P, "unsupported unary operator %q", ex.Op)
		}

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
				childEnv := newEnvironment(nil)
				for i, prm := range f.Params {
					childEnv.declare(prm.Name, argv[i])
				}
				val, err := vm.execBlock(childEnv, f.Body)
				vm.callstack = vm.callstack[:len(vm.callstack)-1]
				if err != nil {
					var re returnErr
					if errors.As(err, &re) {
						return re.value, nil
					}
					return VNull(), convertLoopErr(vm, err)
				}
				return val, nil
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
	case *ArrayLit:
		xs := make([]Value, len(ex.Elems))
		for i, e := range ex.Elems {
			v, err := vm.evalExpr(env, e)
			if err != nil {
				return VNull(), err
			}
			xs[i] = v
		}
		return VArray(xs), nil

	case *ObjectLit:
		m := make(map[string]Value, len(ex.Fields))
		for _, f := range ex.Fields {
			v, err := vm.evalExpr(env, f.Expr)
			if err != nil {
				return VNull(), err
			}
			m[f.Name] = v
		}
		return VObject(m), nil

	case *MemberExpr:
		obj, err := vm.evalExpr(env, ex.Object)
		if err != nil {
			return VNull(), err
		}
		if obj.Kind != ObjectKind {
			return VNull(), vm.rtErr(ex.P, "cannot access field %q on %s", ex.Name, kindName(obj.Kind))
		}
		val, ok := obj.Obj[ex.Name]
		if !ok {
			return VNull(), vm.rtErr(ex.P, "unknown field %q", ex.Name)
		}
		return val, nil

	case *IndexExpr:
		seq, err := vm.evalExpr(env, ex.Seq)
		if err != nil {
			return VNull(), err
		}
		idx, err := vm.evalExpr(env, ex.Index)
		if err != nil {
			return VNull(), err
		}
		switch seq.Kind {
		case ArrayKind:
			i := int(idx.AsInt())
			if i < 0 || i >= len(seq.Arr) {
				return VNull(), vm.rtErr(ex.P, "index out of range")
			}
			return seq.Arr[i], nil
		default:
			return VNull(), vm.rtErr(ex.P, "indexing not supported on %s", kindName(seq.Kind))
		}

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
