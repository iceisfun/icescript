package ast

import (
	"bytes"
	"strings"

	"github.com/iceisfun/icescript/token"
)

type TupleLiteral struct {
	Token    token.Token // '('
	Elements []Expression
}

func (tl *TupleLiteral) expressionNode()      {}
func (tl *TupleLiteral) TokenLiteral() string { return tl.Token.Literal }
func (tl *TupleLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range tl.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("(")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString(")")

	return out.String()
}
