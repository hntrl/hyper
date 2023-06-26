package ast

import (
	"testing"

	"github.com/hntrl/hyper/src/hyper/parser"
	"github.com/hntrl/hyper/src/hyper/tokens"
)

// ArgumentList
// CAN PARSE ARGUMENT LIST
func TestArgumentList(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "{ a: b }, c: d)",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseArgumentList(p)
		},
		expects: &ArgumentList{
			pos: tokens.Position{Line: 1, Column: 1},
			Items: []Node{
				ArgumentObject{
					pos: tokens.Position{Line: 1, Column: 3},
					Items: []ArgumentItem{
						{
							pos: tokens.Position{Line: 1, Column: 3},
							Key: "a",
							Init: TypeExpression{
								pos:        tokens.Position{Line: 1, Column: 6},
								IsArray:    false,
								IsPartial:  false,
								IsOptional: false,
								Selector: Selector{
									pos:     tokens.Position{Line: 1, Column: 6},
									Members: []string{"b"},
								},
							},
						},
					},
				},
				ArgumentItem{
					pos: tokens.Position{Line: 1, Column: 10},
					Key: "c",
					Init: TypeExpression{
						pos:        tokens.Position{Line: 1, Column: 13},
						IsArray:    false,
						IsPartial:  false,
						IsOptional: false,
						Selector: Selector{
							pos:     tokens.Position{Line: 1, Column: 13},
							Members: []string{"d"},
						},
					},
				},
			},
		},
		endingToken: tokens.RPAREN,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN REJECT INVALID ARGUMENT LIST
func TestArgumentListRejectsInvalidArgumentList(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "a: b, c)",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseArgumentList(p)
		},
		expects:      nil,
		expectsError: ExpectedError(tokens.Position{Line: 1, Column: 8}, tokens.COLON, ")"),
	})
	if err != nil {
		t.Error(err)
	}
}

// FunctionBlock
// CAN PARSE FUNCTION BLOCK
func TestFunctionBlock(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `(a: b) c {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseFunctionBlock(p)
		},
		expects: &FunctionBlock{
			Parameters: FunctionParameters{
				Arguments: ArgumentList{
					pos: tokens.Position{Line: 1, Column: 2},
					Items: []Node{
						ArgumentItem{
							pos: tokens.Position{Line: 1, Column: 2},
							Key: "a",
							Init: TypeExpression{
								pos:        tokens.Position{Line: 1, Column: 5},
								IsArray:    false,
								IsPartial:  false,
								IsOptional: false,
								Selector: Selector{
									pos:     tokens.Position{Line: 1, Column: 5},
									Members: []string{"b"},
								},
							},
						},
					},
				},
				ReturnType: &TypeExpression{
					pos:        tokens.Position{Line: 1, Column: 8},
					IsArray:    false,
					IsPartial:  false,
					IsOptional: false,
					Selector: Selector{
						pos:     tokens.Position{Line: 1, Column: 8},
						Members: []string{"c"},
					},
				},
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 11},
				Statements: []BlockStatement{},
			},
		},
		endingToken: tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE FUNCTION BLOCK WITHOUT RETURN TYPE
func TestFunctionBlockWithoutReturnType(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `(a: b) {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseFunctionBlock(p)
		},
		expects: &FunctionBlock{
			Parameters: FunctionParameters{
				Arguments: ArgumentList{
					pos: tokens.Position{Line: 1, Column: 2},
					Items: []Node{
						ArgumentItem{
							pos: tokens.Position{Line: 1, Column: 2},
							Key: "a",
							Init: TypeExpression{
								pos:        tokens.Position{Line: 1, Column: 5},
								IsArray:    false,
								IsPartial:  false,
								IsOptional: false,
								Selector: Selector{
									pos:     tokens.Position{Line: 1, Column: 5},
									Members: []string{"b"},
								},
							},
						},
					},
				},
				ReturnType: nil,
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 8},
				Statements: []BlockStatement{},
			},
		},
		endingToken: tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// Block
// CAN PARSE BLOCK
func TestBlock(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `if (true) {} while (true) {} for (idx, val in true) {} continue break switch (true) {} guard "foo" return "foo" throw "bar" abc.def }`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseBlock(p)
		},
		expects: &Block{
			pos: tokens.Position{Line: 1, Column: 1},
			Statements: []BlockStatement{
				{
					Init: IfStatement{
						pos: tokens.Position{Line: 1, Column: 1},
						Condition: Expression{
							pos: tokens.Position{Line: 1, Column: 5},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 5},
								Value: true,
							},
						},
						Body: Block{
							pos:        tokens.Position{Line: 1, Column: 10},
							Statements: []BlockStatement{},
						},
						Alternate: nil,
					},
				},
				{
					Init: WhileStatement{
						pos: tokens.Position{Line: 1, Column: 10},
						Condition: Expression{
							pos: tokens.Position{Line: 1, Column: 5},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 5},
								Value: true,
							},
						},
						Body: Block{
							pos:        tokens.Position{Line: 1, Column: 10},
							Statements: []BlockStatement{},
						},
					},
				},
				{
					Init: ForStatement{
						pos: tokens.Position{Line: 1, Column: 20},
						Condition: RangeCondition{
							pos:   tokens.Position{Line: 1, Column: 24},
							Index: "idx",
							Value: "val",
							Target: Expression{
								pos: tokens.Position{Line: 1, Column: 5},
								Init: Literal{
									pos:   tokens.Position{Line: 1, Column: 5},
									Value: true,
								},
							},
						},
						Body: Block{
							pos:        tokens.Position{Line: 1, Column: 10},
							Statements: []BlockStatement{},
						},
					},
				},
				{
					Init: ContinueStatement{
						pos: tokens.Position{Line: 1, Column: 38},
					},
				},
				{
					Init: BreakStatement{
						pos: tokens.Position{Line: 1, Column: 48},
					},
				},
				{
					Init: SwitchBlock{
						pos: tokens.Position{Line: 1, Column: 56},
						Target: Expression{
							pos: tokens.Position{Line: 1, Column: 5},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 5},
								Value: true,
							},
						},
						Statements: []SwitchStatement{},
					},
				},
				{
					Init: GuardStatement{
						pos: tokens.Position{Line: 1, Column: 71},
						Init: Expression{
							pos: tokens.Position{Line: 1, Column: 78},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 78},
								Value: "foo",
							},
						},
					},
				},
				{
					Init: ReturnStatement{
						pos: tokens.Position{Line: 1, Column: 1},
						Init: Expression{
							pos: tokens.Position{Line: 1, Column: 8},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 8},
								Value: "foo",
							},
						},
					},
				},
				{
					Init: ThrowStatement{
						pos: tokens.Position{Line: 1, Column: 1},
						Init: Expression{
							pos: tokens.Position{Line: 1, Column: 8},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 8},
								Value: "bar",
							},
						},
					},
				},
				{
					Init: Expression{
						pos: tokens.Position{Line: 1, Column: 13},
						Init: ValueExpression{
							pos: tokens.Position{Line: 1, Column: 1},
							Members: []ValueExpressionMember{
								{
									pos:  tokens.Position{Line: 1, Column: 1},
									Init: "abc",
								},
								{
									pos:  tokens.Position{Line: 1, Column: 5},
									Init: "def",
								},
							},
						},
					},
				},
			},
		},
		expectsError: nil,
		endingToken:  tokens.RCURLY,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN REJECT INVALID BLOCK STATEMENTS
func TestBlockRejectsInvalidBlockStatements(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `+= "foo"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseBlock(p)
		},
		expects:      nil,
		expectsError: ExpectedError(tokens.Position{Line: 1, Column: 1}, tokens.IDENT, "+="),
	})
	if err != nil {
		t.Error(err)
	}
}

// DeclarationStatement
// CAN PARSE DECLARATION STATEMENTS
func TestDeclarationStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `foo := "bar"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseDeclarationStatement(p)
		},
		expects: &DeclarationStatement{
			pos:             tokens.Position{Line: 1, Column: 1},
			Target:          "foo",
			SecondaryTarget: nil,
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 7},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 7},
					Value: "bar",
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

// CAN PARSE DECLARATION STATEMENTS WITH SECONDARY TARGET
func TestDeclarationStatementWithSecondaryTarget(t *testing.T) {
	expectedSecondaryTarget := "bar"
	err := evaluateTest(TestFixture{
		lit: `foo, bar := "baz"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseDeclarationStatement(p)
		},
		expects: &DeclarationStatement{
			pos:             tokens.Position{Line: 1, Column: 1},
			Target:          "foo",
			SecondaryTarget: &expectedSecondaryTarget,
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 12},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 12},
					Value: "baz",
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE DECLARATION STATEMENTS WITH TRY STATEMENT
func TestDeclarationStatementWithTryStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `foo := try "bar"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseDeclarationStatement(p)
		},
		expects: &DeclarationStatement{
			pos:             tokens.Position{Line: 1, Column: 1},
			Target:          "foo",
			SecondaryTarget: nil,
			Init: TryStatement{
				pos: tokens.Position{Line: 1, Column: 7},
				Init: Expression{
					pos: tokens.Position{Line: 1, Column: 11},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 11},
						Value: "bar",
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE DECLARATION STATEMENTS WITH TRY STATEMENT AND SECONDARY TARGET
func TestDeclarationStatementWithTryStatementAndSecondaryTarget(t *testing.T) {
	expectedSecondaryTarget := "baz"
	err := evaluateTest(TestFixture{
		lit: `foo, bar := try "baz"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseDeclarationStatement(p)
		},
		expects: &DeclarationStatement{
			pos:             tokens.Position{Line: 1, Column: 1},
			Target:          "foo",
			SecondaryTarget: &expectedSecondaryTarget,
			Init: TryStatement{
				pos: tokens.Position{Line: 1, Column: 12},
				Init: Expression{
					pos: tokens.Position{Line: 1, Column: 16},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 16},
						Value: "baz",
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

// AssignmentStatement
// CAN CREATE STANDARD ASSIGNMENT
func TestAssignmentStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `abc = 1`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentStatement(p)
		},
		expects: &AssignmentStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Target: AssignmentTargetExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Members: []AssignmentTargetExpressionMember{
					{Init: "abc"},
				},
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
func TestAssignmentStatementInvalidOperator(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `abc != 1`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentStatement(p)
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
func TestAssignmentStatementIntegerOp(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `abc++`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentStatement(p)
		},
		expects: &AssignmentStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Target: AssignmentTargetExpression{
				pos:     tokens.Position{Line: 1, Column: 1},
				Members: []AssignmentTargetExpressionMember{{Init: "abc"}},
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

// CAN CREATE ASSIGNMENT WITH SECONDARY TARGET
func TestAssignmentStatementWithSecondaryTarget(t *testing.T) {
	expectedSecondaryTarget := "bar"
	err := evaluateTest(TestFixture{
		lit: `abc, bar = 1`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentStatement(p)
		},
		expects: &AssignmentStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Target: AssignmentTargetExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Members: []AssignmentTargetExpressionMember{
					{Init: "abc"},
				},
			},
			SecondaryTarget: &expectedSecondaryTarget,
			Operator:        tokens.ASSIGN,
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 9},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 9},
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

// CAN CREATE ASSIGNMENT WITH COMPLEX TARGET AND SECONDARY TARGET
func TestAssignmentStatementWithComplexTargetAndSecondaryTarget(t *testing.T) {
	expectedSecondaryTarget := "bar"
	err := evaluateTest(TestFixture{
		lit: `abc[1].val[3:int('abc')], bar = 1`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentStatement(p)
		},
		expects: &AssignmentStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Target: AssignmentTargetExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Members: []AssignmentTargetExpressionMember{
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
						Init: "val",
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
			SecondaryTarget: &expectedSecondaryTarget,
			Operator:        tokens.ASSIGN,
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 31},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 31},
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

// CAN CREATE ASSIGNMENT WITH TRY STATEMENT
func TestAssignmentStatementWithTryStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `abc := try "bar"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseAssignmentStatement(p)
		},
		expects: &AssignmentStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Target: AssignmentTargetExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Members: []AssignmentTargetExpressionMember{
					{Init: "abc"},
				},
			},
			SecondaryTarget: nil,
			Operator:        tokens.ASSIGN,
			Init: TryStatement{
				pos: tokens.Position{Line: 1, Column: 7},
				Init: Expression{
					pos: tokens.Position{Line: 1, Column: 11},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 11},
						Value: "bar",
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

// AssignmentTargetExpression
// CAN PARSE ASSIGNMENT TARGET EXPRESSION
func TestAssignmentTargetExpression(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "abc[1].val[3:int('abc')]",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseExpression(p)
		},
		expects: &Expression{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: AssignmentTargetExpression{
				pos: tokens.Position{Line: 1, Column: 1},
				Members: []AssignmentTargetExpressionMember{
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
						Init: "val",
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
								Init: AssignmentTargetExpression{
									pos: tokens.Position{Line: 1, Column: 19},
									Members: []AssignmentTargetExpressionMember{
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

// IfStatement
// CAN PARSE IF STATEMENTS
func TestIfStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `if (true) {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseIfStatement(p)
		},
		expects: &IfStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: Expression{
				pos: tokens.Position{Line: 1, Column: 4},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 4},
					Value: true,
				},
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 9},
				Statements: []BlockStatement{},
			},
			Alternate: nil,
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE IF STATEMENTS WITH INLINE BLOCK
func TestIfStatementWithInlineBlock(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `if (true) "foo"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseIfStatement(p)
		},
		expects: &IfStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: Expression{
				pos: tokens.Position{Line: 1, Column: 4},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 4},
					Value: true,
				},
			},
			Body: Block{
				pos: tokens.Position{Line: 1, Column: 9},
				Statements: []BlockStatement{
					{
						Init: Expression{
							pos: tokens.Position{Line: 1, Column: 11},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 11},
								Value: "foo",
							},
						},
					},
				},
			},
			Alternate: nil,
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE IF STATEMENTS WITH ALTERNATE CONDITIONS
func TestIfStatementWithAlternateConditions(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `if (true) {} else if (false) {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseIfStatement(p)
		},
		expects: &IfStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: Expression{
				pos: tokens.Position{Line: 1, Column: 4},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 4},
					Value: true,
				},
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 9},
				Statements: []BlockStatement{},
			},
			Alternate: IfStatement{
				pos: tokens.Position{Line: 1, Column: 16},
				Condition: Expression{
					pos: tokens.Position{Line: 1, Column: 19},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 19},
						Value: false,
					},
				},
				Body: Block{
					pos:        tokens.Position{Line: 1, Column: 25},
					Statements: []BlockStatement{},
				},
				Alternate: nil,
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE IF STATEMENTS WITH ALTERNATE CONDITIONS AND INLINE BLOCKS
func TestIfStatementWithAlternateConditionsAndInlineBlocks(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `if (true) "foo" else if (false) "bar"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseIfStatement(p)
		},
		expects: &IfStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: Expression{
				pos: tokens.Position{Line: 1, Column: 4},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 4},
					Value: true,
				},
			},
			Body: Block{
				pos: tokens.Position{Line: 1, Column: 9},
				Statements: []BlockStatement{
					{
						Init: Expression{
							pos: tokens.Position{Line: 1, Column: 11},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 11},
								Value: "foo",
							},
						},
					},
				},
			},
			Alternate: IfStatement{
				pos: tokens.Position{Line: 1, Column: 18},
				Condition: Expression{
					pos: tokens.Position{Line: 1, Column: 21},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 21},
						Value: false,
					},
				},
				Body: Block{
					pos: tokens.Position{Line: 1, Column: 27},
					Statements: []BlockStatement{
						{
							Init: Expression{
								pos: tokens.Position{Line: 1, Column: 29},
								Init: Literal{
									pos:   tokens.Position{Line: 1, Column: 29},
									Value: "bar",
								},
							},
						},
					},
				},
				Alternate: nil,
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE IF STATEMENT WITH ELSE BLOCK
func TestIfStatementWithElseBlock(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `if (true) {} else {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseIfStatement(p)
		},
		expects: &IfStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: Expression{
				pos: tokens.Position{Line: 1, Column: 4},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 4},
					Value: true,
				},
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 9},
				Statements: []BlockStatement{},
			},
			Alternate: Block{
				pos:        tokens.Position{Line: 1, Column: 25},
				Statements: []BlockStatement{},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// WhileStatement
// CAN PARSE WHILE STATEMENTS
func TestWhileStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `while (true) {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseWhileStatement(p)
		},
		expects: &WhileStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: Expression{
				pos: tokens.Position{Line: 1, Column: 7},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 7},
					Value: true,
				},
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 12},
				Statements: []BlockStatement{},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE WHILE STATEMENTS WITH INLINE BLOCK
func TestWhileStatementWithInlineBlock(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `while (true) "foo"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseWhileStatement(p)
		},
		expects: &WhileStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: Expression{
				pos: tokens.Position{Line: 1, Column: 7},
				Init: Literal{
					pos:   tokens.Position{Line: 1, Column: 7},
					Value: true,
				},
			},
			Body: Block{
				pos: tokens.Position{Line: 1, Column: 12},
				Statements: []BlockStatement{
					{
						Init: Expression{
							pos: tokens.Position{Line: 1, Column: 14},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 14},
								Value: "foo",
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

// ForStatement
// CAN PARSE FOR STATEMENT WITH RANGE CONDITION
func TestForStatementWithRangeCondition(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `for (idx, val in "foo") {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseForStatement(p)
		},
		expects: &ForStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: RangeCondition{
				pos:   tokens.Position{Line: 1, Column: 8},
				Index: "idx",
				Value: "val",
				Target: Expression{
					pos: tokens.Position{Line: 1, Column: 18},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 18},
						Value: "foo",
					},
				},
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 17},
				Statements: []BlockStatement{},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE FOR STATEMENT WITH AGGREGATE FOR CONDITION
func TestForStatementWithAggregateForCondition(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `for (idx := 0; idx <= 5; idx++) {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseForStatement(p)
		},
		expects: &ForStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: ForCondition{
				pos: tokens.Position{Line: 1, Column: 8},
				Init: &DeclarationStatement{
					pos:  tokens.Position{Line: 1, Column: 8},
					Name: "idx",
					Init: Expression{
						pos: tokens.Position{Line: 1, Column: 13},
						Init: Literal{
							pos:   tokens.Position{Line: 1, Column: 13},
							Value: int64(0),
						},
					},
				},
				Condition: Expression{
					pos: tokens.Position{Line: 1, Column: 18},
					Init: BinaryExpression{
						pos: tokens.Position{Line: 1, Column: 18},
						Left: Expression{
							pos: tokens.Position{Line: 1, Column: 18},
							Init: ValueExpression{
								pos: tokens.Position{Line: 1, Column: 18},
								Members: []ValueExpressionMember{
									{
										pos:  tokens.Position{Line: 1, Column: 18},
										Init: "idx",
									},
								},
							},
						},
						Operator: tokens.LESS_EQUAL,
						Right: Expression{
							pos: tokens.Position{Line: 1, Column: 23},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 23},
								Value: int64(5),
							},
						},
					},
				},
				Update: AssignmentStatement{
					pos: tokens.Position{Line: 1, Column: 28},
					Target: AssignmentTargetExpression{
						pos: tokens.Position{Line: 1, Column: 28},
						Members: []AssignmentTargetExpressionMember{
							{Init: "idx"},
						},
					},
					Operator: tokens.ADD,
					Init: Expression{
						pos: tokens.Position{Line: 1, Column: 31},
						Init: Literal{
							pos:   tokens.Position{Line: 1, Column: 31},
							Value: int64(1),
						},
					},
				},
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 17},
				Statements: []BlockStatement{},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN PARSE FOR STATEMENT WITH AGGREGATE FOR CONDITION WITHOUT DECLARATION
func TestForStatementWithAggregateForConditionWithoutDeclaration(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `for (idx <= 5; true) {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseForStatement(p)
		},
		expects: &ForStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Condition: ForCondition{
				pos:  tokens.Position{Line: 1, Column: 8},
				Init: nil,
				Condition: Expression{
					pos: tokens.Position{Line: 1, Column: 17},
					Init: BinaryExpression{
						pos: tokens.Position{Line: 1, Column: 17},
						Left: Expression{
							pos: tokens.Position{Line: 1, Column: 17},
							Init: ValueExpression{
								pos: tokens.Position{Line: 1, Column: 17},
								Members: []ValueExpressionMember{
									{
										pos:  tokens.Position{Line: 1, Column: 17},
										Init: "idx",
									},
								},
							},
						},
						Operator: tokens.LESS_EQUAL,
						Right: Expression{
							pos: tokens.Position{Line: 1, Column: 22},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 22},
								Value: int64(5),
							},
						},
					},
				},
				Update: Expression{
					pos: tokens.Position{Line: 1, Column: 27},
					Init: Literal{
						pos:   tokens.Position{Line: 1, Column: 27},
						Value: true,
					},
				},
			},
			Body: Block{
				pos:        tokens.Position{Line: 1, Column: 17},
				Statements: []BlockStatement{},
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// SwitchBlock
// CAN PARSE SWITCH BLOCK
func TestSwitchBlock(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `switch (foo) { case "bar": break default: continue }`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseSwitchBlock(p)
		},
		expects: &SwitchBlock{
			pos: tokens.Position{Line: 1, Column: 1},
			Target: Expression{
				pos: tokens.Position{Line: 1, Column: 9},
				Init: ValueExpression{
					pos: tokens.Position{Line: 1, Column: 9},
					Members: []ValueExpressionMember{
						{
							pos:  tokens.Position{Line: 1, Column: 9},
							Init: "foo",
						},
					},
				},
			},
			Statements: []SwitchStatement{
				{
					pos: tokens.Position{Line: 1, Column: 14},
					Condition: &Expression{
						pos: tokens.Position{Line: 1, Column: 20},
						Init: Literal{
							pos:   tokens.Position{Line: 1, Column: 20},
							Value: "bar",
						},
					},
					IsDefault: false,
					Body: Block{
						pos: tokens.Position{Line: 1, Column: 26},
						Statements: []BlockStatement{
							{
								Init: BreakStatement{
									pos: tokens.Position{Line: 1, Column: 27},
								},
							},
						},
					},
				},
				{
					pos:       tokens.Position{Line: 1, Column: 34},
					Condition: nil,
					IsDefault: true,
					Body: Block{
						pos: tokens.Position{Line: 1, Column: 41},
						Statements: []BlockStatement{
							{
								Init: ContinueStatement{
									pos: tokens.Position{Line: 1, Column: 42},
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

// GuardStatement
// CAN PARSE GUARD STATEMENT
func TestGuardStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `guard foo`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseGuardStatement(p)
		},
		expects: &GuardStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 7},
				Init: ValueExpression{
					pos: tokens.Position{Line: 1, Column: 7},
					Members: []ValueExpressionMember{
						{
							pos:  tokens.Position{Line: 1, Column: 7},
							Init: "foo",
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

// ReturnStatement
// CAN PARSE RETURN STATEMENT
func TestReturnStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `return foo`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseReturnStatement(p)
		},
		expects: &ReturnStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 7},
				Init: ValueExpression{
					pos: tokens.Position{Line: 1, Column: 7},
					Members: []ValueExpressionMember{
						{
							pos:  tokens.Position{Line: 1, Column: 7},
							Init: "foo",
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

// ThrowStatement
// CAN PARSE THROW STATEMENT
func TestThrowStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `throw foo`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseThrowStatement(p)
		},
		expects: &ThrowStatement{
			pos: tokens.Position{Line: 1, Column: 1},
			Init: Expression{
				pos: tokens.Position{Line: 1, Column: 7},
				Init: ValueExpression{
					pos: tokens.Position{Line: 1, Column: 7},
					Members: []ValueExpressionMember{
						{
							pos:  tokens.Position{Line: 1, Column: 7},
							Init: "foo",
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
