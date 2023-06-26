package ast

import (
	"testing"

	"github.com/hntrl/hyper/src/hyper/parser"
	"github.com/hntrl/hyper/src/hyper/tokens"
)

// ObjectPattern
// CAN PARSE OBJECT PATTERN
func TestObjectPattern(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `{ foo: "bar" }`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseObjectPattern(p)
		},
		expects: &ObjectPattern{
			pos: tokens.Position{Line: 1, Column: 1},
			Properties: PropertyList{
				Property{
					pos: tokens.Position{Line: 1, Column: 2},
					Key: "foo",
					Init: Expression{
						pos: tokens.Position{Line: 1, Column: 7},
						Init: Literal{
							pos:   tokens.Position{Line: 1, Column: 7},
							Value: "bar",
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

// CAN PARSE EMPTY OBJECT PATTERN
func TestEmptyObjectPattern(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `{}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseObjectPattern(p)
		},
		expects: &ObjectPattern{
			pos:        tokens.Position{Line: 1, Column: 1},
			Properties: PropertyList{},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE OBJECT PATTERN WITH SPREAD ELEMENT
func TestObjectPatternWithSpreadElement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `{ ...foo }`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseObjectPattern(p)
		},
		expects: &ObjectPattern{
			pos: tokens.Position{Line: 1, Column: 1},
			Properties: PropertyList{
				SpreadElement{
					pos: tokens.Position{Line: 1, Column: 2},
					Init: Expression{
						pos: tokens.Position{Line: 1, Column: 6},
						Init: ValueExpression{
							pos: tokens.Position{Line: 1, Column: 6},
							Members: []ValueExpressionMember{
								{
									pos:  tokens.Position{Line: 1, Column: 6},
									Init: "foo",
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
