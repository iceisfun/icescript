package icescript

import "testing"

func TestParseIfAndFor(t *testing.T) {
	src := `
func demo() int {
  var total = 0
  var nums = [1, 2, 3]
  for n in nums {
    if (n % 2 == 0) {
      total = total + n
    }
  }
  return total
}
`

	parser := NewParser(New(src))
	program, errs := parser.ParseProgram()
	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	fn, ok := program.Funcs["demo"]
	if !ok {
		t.Fatalf("function 'demo' not found")
	}

	if len(fn.Body.Stmts) != 4 {
		t.Fatalf("expected 5 statements in demo body, got %d", len(fn.Body.Stmts))
	}

	if _, ok := fn.Body.Stmts[0].(*VarStmt); !ok {
		t.Fatalf("expected first statement to be VarStmt, got %T", fn.Body.Stmts[0])
	}

	forStmt, ok := fn.Body.Stmts[2].(*ForInStmt)
	if !ok {
		t.Fatalf("expected third statement to be ForInStmt, got %T", fn.Body.Stmts[2])
	}

	if forStmt.VarName != "n" {
		t.Fatalf("expected loop variable 'n', got %s", forStmt.VarName)
	}

	ifStmt, ok := forStmt.Body.Stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected loop body to start with IfStmt, got %T", forStmt.Body.Stmts[0])
	}

	assign, ok := ifStmt.Then.Stmts[0].(*AssignStmt)
	if !ok {
		t.Fatalf("expected assignment in if body, got %T", ifStmt.Then.Stmts[0])
	}

	if _, ok := assign.Target.(*Ident); !ok {
		t.Fatalf("expected assignment target to be Ident, got %T", assign.Target)
	}
}

func TestParseMemberAssignment(t *testing.T) {
	src := `
func update() {
  Player.position.x = Player.position.x + 1
  return null
}
`

	parser := NewParser(New(src))
	program, errs := parser.ParseProgram()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	fn := program.Funcs["update"]
	if fn == nil {
		t.Fatalf("function 'update' missing")
	}

	assign, ok := fn.Body.Stmts[0].(*AssignStmt)
	if !ok {
		t.Fatalf("expected first statement to be AssignStmt, got %T", fn.Body.Stmts[0])
	}

	member, ok := assign.Target.(*MemberExpr)
	if !ok {
		t.Fatalf("expected member expression target, got %T", assign.Target)
	}

	if member.Name != "x" {
		t.Fatalf("expected field 'x', got %s", member.Name)
	}
}
