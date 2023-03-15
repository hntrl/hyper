package nodes

import (
	"testing"

	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"
)

// TypeExpression
// CAN CREATE TYPE EXPRESSION
func TestTypeExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "Foo",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseTypeExpression(p)
		},
		expects: &TypeExpression{
			pos:        tokens.Position{Line: 1, Column: 6},
			IsArray:    false,
			IsPartial:  false,
			IsOptional: false,
			Selector: Selector{
				pos:     tokens.Position{Line: 1, Column: 1},
				Members: []string{"Foo"},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE OPTIONAL TYPE EXPRESSION
func TestTypeExpressionOptional(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "Foo?",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseTypeExpression(p)
		},
		expects: &TypeExpression{
			pos:        tokens.Position{Line: 1, Column: 6},
			IsArray:    false,
			IsPartial:  false,
			IsOptional: true,
			Selector: Selector{
				pos:     tokens.Position{Line: 1, Column: 1},
				Members: []string{"Foo"},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE ARRAY TYPE EXPRESSION
func TestTypeExpressionArray(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "[]Foo",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseTypeExpression(p)
		},
		expects: &TypeExpression{
			pos:        tokens.Position{Line: 1, Column: 1},
			IsArray:    true,
			IsPartial:  false,
			IsOptional: false,
			Selector: Selector{
				pos:     tokens.Position{Line: 1, Column: 3},
				Members: []string{"Foo"},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE PARTIAL TYPE EXPRESSION
func TestTypeExpressionPartial(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "Partial<Foo>",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseTypeExpression(p)
		},
		expects: &TypeExpression{
			pos:        tokens.Position{Line: 1, Column: 1},
			IsArray:    false,
			IsPartial:  true,
			IsOptional: false,
			Selector: Selector{
				pos:     tokens.Position{Line: 1, Column: 3},
				Members: []string{"Foo"},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// Expression
// CAN PARSE ARRAY EXPRESSION
func TestArrayExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `[]String{"foo", "bar"}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: ArrayExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Init: TypeExpression{
					pos:        tokens.Position{Line: 1, Column: 3},
					IsArray:    false,
					IsPartial:  false,
					IsOptional: false,
					Selector: Selector{
						pos:     tokens.Position{Line: 1, Column: 3},
						Members: []string{"String"},
					},
				},
				Elements: []Expression{
					{
						pos: tokens.Position{Line: 1, Column: 2},
						Init: Literal{
							pos:   tokens.Position{Line: 1, Column: 2},
							Value: "foo",
						},
					},
					{
						pos: tokens.Position{Line: 1, Column: 8},
						Init: Literal{
							pos:   tokens.Position{Line: 1, Column: 8},
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

// CAN PARSE INSTANCE EXPRESSION
func TestInstanceExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `Foo{}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: InstanceExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Selector: Selector{
					pos:     tokens.Position{Line: 1, Column: 1},
					Members: []string{"Foo"},
				},
				Properties: PropertyList{},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE UNARY EXPRESSION
func TestUnaryExpression(t *testing.T) {
	unaryFixture := func(tok tokens.Token) TestFixture {
		return TestFixture{
			lit: tok.String() + "true",
			parseFn: func(p *parser.Parser) (Node, error) {
				return ParseExpression(p)
			},
			expects: &Expression{
				pos: tokens.Position{Line: 1, Column: 1},
				Init: UnaryExpression{
					pos:      tokens.Position{Line: 1, Column: 1},
					Operator: tok,
					Init: Expression{
						pos: tokens.Position{Line: 1, Column: 2},
						Init: Literal{
							pos:   tokens.Position{Line: 1, Column: 2},
							Value: true,
						},
					},
				},
			},
			expectsError: nil,
			endingToken:  tokens.EOF,
		}
	}
	t.Run("!", func(t *testing.T) {
		err := evaluateTest(unaryFixture(tokens.NOT))
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("+", func(t *testing.T) {
		err := evaluateTest(unaryFixture(tokens.SUB))
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("-", func(t *testing.T) {
		err := evaluateTest(unaryFixture(tokens.SUB))
		if err != nil {
			t.Error(err)
		}
	})
}

// CAN PARSE BINARY EXPRESSION
func TestBinaryExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "123.456 ** 789",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: BinaryExpression{
				pos:      tokens.Position{Line: 1, Column: 1},
				Operator: tokens.PWR,
				Left: Expression{
					pos: tokens.Position{Line: 1, Column: 1},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 1},
						Value: float64(123.456),
					},
				},
				Right: Expression{
					pos: tokens.Position{Line: 1, Column: 2},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 2},
						Value: int64(789),
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

// CAN PARSE FUNCTION EXPRESSION
func TestFunctionExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `func() {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: FunctionExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Body: FunctionBlock{
					Parameters: FunctionParameters{
						Arguments: ArgumentList{
							pos:   tokens.Position{Line: 1, Column: 11},
							Items: make([]Node, 0),
						},
						ReturnType: nil,
					},
					Body: Block{
						pos:        tokens.Position{Line: 1, Column: 12},
						Statements: []BlockStatement{},
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

// CAN PARSE VALUE EXPRESSION
func TestValueExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "abc[1].fn()[3:int('abc')]",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: ValueExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Members: []ValueExpressionMember{
					{
						pos:  tokens.Position{Line: 1, Column: 1},
						Init: "abc",
					},
					{
						pos: tokens.Position{Line: 1, Column: 4},
						Init: IndexExpression{
							pos: tokens.Position{Line: 1, Column: 4},
							Left: &Expression{
								pos: tokens.Position{Line: 1, Column: 5},
								Init: Literal{
									pos:   tokens.Position{Line: 1, Column: 5},
									Value: int64(1),
								},
							},
							IsRange: false,
							Right:   nil,
						},
					},
					{
						pos:  tokens.Position{Line: 1, Column: 7},
						Init: "fn",
					},
					{
						pos: tokens.Position{Line: 1, Column: 10},
						Init: CallExpression{
							pos:       tokens.Position{Line: 1, Column: 10},
							Arguments: []Expression{},
						},
					},
					{
						pos: tokens.Position{Line: 1, Column: 13},
						Init: IndexExpression{
							pos: tokens.Position{Line: 1, Column: 13},
							Left: &Expression{
								pos: tokens.Position{Line: 1, Column: 14},
								Init: Literal{
									pos:   tokens.Position{Line: 1, Column: 14},
									Value: int64(3),
								},
							},
							IsRange: true,
							Right: &Expression{
								pos: tokens.Position{Line: 1, Column: 19},
								Init: ValueExpression{
									pos: tokens.Position{Line: 1, Column: 19},
									Members: []ValueExpressionMember{
										{
											pos:  tokens.Position{Line: 1, Column: 19},
											Init: "int",
										},
										{
											pos: tokens.Position{Line: 1, Column: 22},
											Init: CallExpression{
												pos: tokens.Position{Line: 1, Column: 22},
												Arguments: []Expression{
													{
														pos: tokens.Position{Line: 1, Column: 23},
														Init: Literal{
															pos:   tokens.Position{Line: 1, Column: 23},
															Value: "abc",
														},
													},
												},
											},
										},
									},
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

// REJECTS CALL EXPRESSION WITH INVALID ARGUMENTS
func TestCallExpressionInvalidArguments(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "fn(1 2)",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects:      nil,
		expectsError: ExpectedError(tokens.Position{Line: 1, Column: 6}, tokens.COMMA, "2"),
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE EMBEDDED EXPRESSION
func TestEmbeddedExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `('abc')`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 1},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 2},
					Value: "abc",
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

// AssignmentExpression
// CAN CREATE STANDARD ASSIGNMENT
func TestAssignmentExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `abc = 1`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentExpression(p)
		},
		expects: &AssignmentExpression{
			pos: tokens.Position{Line: 1, Column: 1},
			Name: Selector{
				pos:     tokens.Position{Line: 1, Column: 1},
				Members: []string{"abc"},
			},
			Operator: tokens.ASSIGN,
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 5},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 5},
					Value: int64(1),
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

// CAN REJECT ASSIGNMENT WITH INVALID OPERATOR
func TestAssignmentExpressionInvalidOperator(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `abc != 1`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentExpression(p)
		},
		expects:      nil,
		expectsError: ExpectedError(tokens.Position{Line: 1, Column: 5}, tokens.ILLEGAL, "!="),
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE INTEGER OP ASSIGNMENT
func TestAssignmentExpressionIntegerOp(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `abc++`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentExpression(p)
		},
		expects: &AssignmentExpression{
			pos: tokens.Position{Line: 1, Column: 1},
			Name: Selector{
				pos:     tokens.Position{Line: 1, Column: 1},
				Members: []string{"abc"},
			},
			Operator: tokens.ADD,
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 6},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 6},
					Value: int64(1),
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
