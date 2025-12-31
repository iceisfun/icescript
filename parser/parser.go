package parser

import (
	"fmt"
	"strconv"

	"github.com/iceisfun/icescript/ast"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/token"
)

const (
	_ int = iota
	LOWEST
	ASSIGN      // =
	LOGICAL_OR  // ||
	LOGICAL_AND // &&
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MOD:      PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.ASSIGN:   ASSIGN,
	token.AND:      LOGICAL_AND,
	token.OR:       LOGICAL_OR,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []token.ScriptError // Use structured errors

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []token.ScriptError{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.LBRACE, p.parseMapLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.NULL, p.parseNullLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) Errors() []string {
	// Compatibility: Convert ScriptError to string format
	var errs []string
	for _, e := range p.errors {
		errs = append(errs, e.Error())
	}
	return errs
}

func (p *Parser) StructuredErrors() []token.ScriptError {
	return p.errors
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.VAR:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FOR:
		return p.parseForStatement() // Placeholder for loop parsing
	case token.FUNCTION:
		// Check for function declaration: func name() {}
		if p.peekTokenIs(token.IDENT) {
			return p.parseFunctionDeclaration()
		}
		// Otherwise, it's a function literal expression statement
		return p.parseExpressionStatement()
	case token.SEMICOLON:
		return nil
	default:
		if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN_DECLARE) {
			return p.parseShortVarDeclaration()
		}
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseShortVarDeclaration() *ast.ShortVarDeclaration {
	stmt := &ast.ShortVarDeclaration{Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}}

	p.nextToken()           // consume IDENT
	stmt.Token = p.curToken // :=

	p.nextToken() // consume :=

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
		// Implicit return null
		stmt.ReturnValue = nil
		return stmt
	}

	if p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) {
		// Implicit return null at end of block
		stmt.ReturnValue = nil
		return stmt
	}

	// Also check if next token starts a new statement that definitely isn't an expression
	if p.peekTokenIs(token.VAR) || p.peekTokenIs(token.RETURN) {
		stmt.ReturnValue = nil
		return stmt
	}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, token.ScriptError{
			Kind:    token.ErrorKindParse,
			Message: msg,
			Line:    p.curToken.Line,
		})
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, token.ScriptError{
			Kind:    token.ErrorKindParse,
			Message: msg,
			Line:    p.curToken.Line,
		})
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// Implementing helper methods first

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if p.peekTokenIs(end) {
			break
		}
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	startToken := p.curToken
	p.nextToken()

	// Check if this is a slice with no start index i.e. [:
	if p.curTokenIs(token.COLON) {
		p.nextToken() // move past COLON
		sliceExp := &ast.SliceExpression{Token: startToken, Left: left}

		// check if there is an end index
		if p.curTokenIs(token.RBRACKET) {
			// No end index, we are at ]
			return sliceExp
		}

		sliceExp.End = p.parseExpression(LOWEST)

		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		return sliceExp
	}

	exp := p.parseExpression(LOWEST)

	if p.peekTokenIs(token.COLON) {
		p.nextToken() // move to COLON
		p.nextToken() // move past COLON
		sliceExp := &ast.SliceExpression{Token: startToken, Left: left, Start: exp}

		if p.curTokenIs(token.RBRACKET) {
			return sliceExp
		}

		sliceExp.End = p.parseExpression(LOWEST)

		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		return sliceExp
	}

	indexExp := &ast.IndexExpression{Token: startToken, Left: left, Index: exp}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return indexExp
}

func (p *Parser) parseMapLiteral() ast.Expression {
	hash := &ast.MapLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		if p.curTokenIs(token.SEMICOLON) {
			continue
		}

		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) {
			if !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.SEMICOLON) {
				return nil
			}
			p.nextToken()
			for p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

// For Loop Support
// Simple C-style support for now: for init; cond; post { body }
func (p *Parser) parseForStatement() ast.Statement {
	stmt := &ast.ForStatement{Token: p.curToken}
	p.nextToken() // Consume FOR

	// Infinite loop: for { ... }
	if p.curTokenIs(token.LBRACE) {
		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	// Helper to check if we are looking at a C-style loop
	// If we just see an expression, it might be the condition (for x < 10)
	// or the init (for x = 0; ...)

	if p.curTokenIs(token.VAR) {
		// usage: for var i = 0; i < 10; i = i + 1 { ... }
		stmt.Init = p.parseLetStatement()
		// parseLetStatement might consume the semicolon if present.
		// If it did, curToken is SEMICOLON.
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken() // move past ;
		}
		// now parse condition
		stmt.Condition = p.parseExpression(LOWEST)

		if !p.expectPeek(token.SEMICOLON) {
			return nil
		}
		p.nextToken() // move past ; (expectPeek moves to it, nextToken moves past it?)
		// expectPeek: if peek is ;, nextToken() (cur becomes ;), return true.
		// So curToken is ;. We need one more nextToken to move to start of post.

		// Parse Post
		postExp := p.parseExpression(LOWEST)
		stmt.Post = &ast.ExpressionStatement{Expression: postExp}

		if !p.expectPeek(token.LBRACE) {
			return nil
		}
	} else if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN_DECLARE) {
		// usage: for i := 0; i < 10; i = i + 1 { ... }
		stmt.Init = p.parseShortVarDeclaration()

		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}

		stmt.Condition = p.parseExpression(LOWEST)

		if !p.expectPeek(token.SEMICOLON) {
			return nil
		}
		p.nextToken()

		postExp := p.parseExpression(LOWEST)
		stmt.Post = &ast.ExpressionStatement{Expression: postExp}

		if !p.expectPeek(token.LBRACE) {
			return nil
		}
	} else {

		// Could be init (expr) or condition
		// We parse an expression.
		expr := p.parseExpression(LOWEST)

		if p.peekTokenIs(token.SEMICOLON) {
			// It was init
			p.nextToken() // curToken becomes ;
			p.nextToken() // curToken becomes start of condition

			stmt.Init = &ast.ExpressionStatement{Expression: expr, Token: token.Token{Literal: expr.TokenLiteral(), Type: token.IDENT}} // Type is guess, mostly purely for string repr
			// Note: AST ExpressionStatement Token field is problematic here since we don't have the original start token handy easily if we didn't save it.
			// But creating a wrapper is fine.

			// Parse Condition
			if !p.curTokenIs(token.SEMICOLON) {
				stmt.Condition = p.parseExpression(LOWEST)
			}

			if !p.expectPeek(token.SEMICOLON) {
				return nil
			}
			p.nextToken() // move past ;

			// Parse Post
			if !p.curTokenIs(token.LBRACE) {
				postExp := p.parseExpression(LOWEST)
				stmt.Post = &ast.ExpressionStatement{Expression: postExp}
			}

			if !p.expectPeek(token.LBRACE) {
				// The post expression parsing loop might have stopped at LBRACE
				// parseExpression loop checks peek.
				// If peek is not semicolon... wait
				// parseExpression precedence check loop stops if peek is token with lower precedence?
				// token.LBRACE precedence? It's not in map -> LOWEST.
				// So parseExpression stops before {.
				if !p.expectPeek(token.LBRACE) {
					return nil
				}
			}
		} else {
			// It was condition
			stmt.Condition = expr
			if !p.expectPeek(token.LBRACE) {
				return nil
			}
		}
	}
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseFunctionDeclaration() ast.Statement {
	// Syntactic sugar: func name(...) { ... }  => var name = func(...) { ... }
	stmt := &ast.LetStatement{Token: token.Token{Type: token.VAR, Literal: "var"}}

	p.nextToken() // consume FUNC

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Function literal part
	lit := &ast.FunctionLiteral{
		Token: token.Token{Type: token.FUNCTION, Literal: "func"},
		Name:  stmt.Name.Value,
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	stmt.Value = lit

	return stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, token.ScriptError{
		Kind:    token.ErrorKindParse,
		Message: msg,
		Line:    p.peekToken.Line,
	})
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, token.ScriptError{
		Kind:    token.ErrorKindParse,
		Message: msg,
		Line:    p.curToken.Line,
	})
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	stmt := &ast.AssignExpression{Token: p.curToken}

	if n, ok := left.(*ast.Identifier); ok {
		stmt.Name = n
	} else {
		p.errors = append(p.errors, token.ScriptError{
			Kind:    token.ErrorKindParse,
			Message: fmt.Sprintf("expected identifier on left side of assignment, got %T", left),
			Line:    p.curToken.Line,
		})
		return nil
	}
	// check if the current token is an assign, if so, consume it
	// Wait, parseAssignExpression is called when we encounter the ASSIGN token as an infix operator.
	// So p.curToken IS ALREADY the ASSIGN token.

	precedence := p.curPrecedence()            // ASSIGN precedence
	p.nextToken()                              // move to start of value expression
	stmt.Value = p.parseExpression(precedence) // was LOWEST, but let's respect precedence usually?
	return stmt
}
