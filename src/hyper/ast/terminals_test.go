package ast

import (
	"testing"

	"github.com/hntrl/hyper/src/hyper/parser"
	"github.com/hntrl/hyper/src/hyper/tokens"
)

// Selector
// CAN PARSE SELECTOR
func TestSelector(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "a.b.c",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseSelector(p)
		},
		expects: &Selector{
			pos:     tokens.Position{Line: 1, Column: 1},
			Members: []string{"a", "b", "c"},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE SELECTOR WITH SINGLE MEMBER
func TestSelectorWithSingleMember(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "a",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseSelector(p)
		},
		expects: &Selector{
			pos:     tokens.Position{Line: 1, Column: 1},
			Members: []string{"a"},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

// Literal
// CAN PARSE STRING LITERAL
func TestStringLiteral(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `"foo"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: Literal{
				pos:   tokens.Position{Line: 1, Column: 1},
				Value: "foo",
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE NUMBER LITERAL
func TestNumberLiteral(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `123`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: Literal{
				pos:   tokens.Position{Line: 1, Column: 1},
				Value: int64(123),
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE FLOAT LITERAL
func TestFloatLiteral(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `123.456`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: Literal{
				pos:   tokens.Position{Line: 1, Column: 1},
				Value: float64(123.456),
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE BOOLEAN LITERAL
func TestBooleanLiteral(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `true`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: Literal{
				pos:   tokens.Position{Line: 1, Column: 1},
				Value: true,
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// ParseTemplateLiteral
// CAN PARSE TEMPLATE LITERAL
func TestTemplateLiteral(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "`foo`",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &TemplateLiteral{
			pos: tokens.Position{Line: 1, Column: 1},
			Parts: []interface{}{
				"foo",
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE TEMPLATE LITERAL WITH EXPRESSION
func TestTemplateLiteralWithExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "`foo {bar}`",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: TemplateLiteral{
				pos: tokens.Position{Line: 1, Column: 1},
				Parts: []interface{}{
					"foo ",
					Expression{
						pos: tokens.Position{Line: 1, Column: 6},
						Init: ValueExpression{
							pos: tokens.Position{Line: 1, Column: 6},
							Members: []ValueExpressionMember{
								{
									pos:  tokens.Position{Line: 1, Column: 6},
									Init: "bar",
								},
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

// CAN PARSE TEMPLATE LITERAL WITH ESCAPED EXPRESSION
func TestTemplateLiteralWithEscapedExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "`foo \\{bar}`",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: TemplateLiteral{
				pos: tokens.Position{Line: 1, Column: 1},
				Parts: []interface{}{
					"foo {bar}",
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

// CAN PARSE TEMPLATE LITERAL WITH ESCAPED BACKTICK
func TestTemplateLiteralWithEscapedBacktick(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "`foo \\`bar`",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: TemplateLiteral{
				pos: tokens.Position{Line: 1, Column: 1},
				Parts: []interface{}{
					"foo `bar",
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

// CAN PARSE TEMPLATE LITERAL WITH MULTIPLE EXPRESSIONS
func TestTemplateLiteralWithMultipleExpressions(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "`foo {bar} {baz}`",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: TemplateLiteral{
				pos: tokens.Position{Line: 1, Column: 1},
				Parts: []interface{}{
					"foo ",
					Expression{
						pos: tokens.Position{Line: 1, Column: 6},
						Init: ValueExpression{
							pos: tokens.Position{Line: 1, Column: 6},
							Members: []ValueExpressionMember{
								{
									pos:  tokens.Position{Line: 1, Column: 6},
									Init: "bar",
								},
							},
						},
					},
					" ",
					Expression{
						pos: tokens.Position{Line: 1, Column: 13},
						Init: ValueExpression{
							pos: tokens.Position{Line: 1, Column: 13},
							Members: []ValueExpressionMember{
								{
									pos:  tokens.Position{Line: 1, Column: 13},
									Init: "baz",
								},
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
