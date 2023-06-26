package ast

import (
	"testing"

	"github.com/hntrl/hyper/src/hyper/parser"
	"github.com/hntrl/hyper/src/hyper/tokens"
)

// FieldStatement
// CAN PARSE FIELD STATEMENT WITH FIELD ASSIGNMENT EXPRESSION
func TestFieldStatementWithFieldAssignmentExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "a = b",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseFieldStatement(p)
		},
		expects: &FieldStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: FieldAssignmentExpression{
				pos:  tokens.Position{Line: 1, Column: 5},
				Name: "a",
				Init: Expression{
					pos: tokens.Position{Line: 1, Column: 9},
					Init: ValueExpression{
						pos: tokens.Position{Line: 1, Column: 9},
						Members: []ValueExpressionMember{
							{
								pos:  tokens.Position{Line: 1, Column: 9},
								Init: "b",
							},
						},
					},
				},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE FIELD STATEMENT WITH ENUM EXPRESSION
func TestFieldStatementWithEnumExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `foo "bar"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseFieldStatement(p)
		},
		expects: &FieldStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: EnumExpression{
				pos:  tokens.Position{Line: 1, Column: 5},
				Name: "foo",
				Init: "bar",
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE FIELD STATEMENT WITH FIELD EXPRESSION
func TestFieldStatementWithTypeStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `foo Bar`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseFieldStatement(p)
		},
		expects: &FieldStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: FieldExpression{
				pos:  tokens.Position{Line: 1, Column: 5},
				Name: "foo",
				Init: TypeExpression{
					pos:        tokens.Position{Line: 1, Column: 9},
					IsArray:    false,
					IsPartial:  false,
					IsOptional: false,
					Selector: Selector{
						pos:     tokens.Position{Line: 1, Column: 9},
						Members: []string{"Bar"},
					},
				},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE FIELD STATEMENT WITH COMMENT
func TestFieldStatementWithComment(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `/* foo */ bar "baz"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseFieldStatement(p)
		},
		expects: &FieldStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: EnumExpression{
				pos:  tokens.Position{Line: 1, Column: 12},
				Name: "bar",
				Init: "baz",
			},
			Comment: " foo ",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}
