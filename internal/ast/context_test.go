package ast

import (
	"testing"

	"github.com/hntrl/hyper/internal/parser"
	"github.com/hntrl/hyper/internal/tokens"
)

// Context
// CAN CREATE CONTEXT
func TestContext(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "context foo { use \"test\" test bar { } func (foo) bar() { } //comment\nfoo bar() {} func baz() { } }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContext(p)
		},
		expects: &Context{
			pos:  tokens.Position{Line: 1, Column: 9},
			Name: "foo",
			Remotes: []UseStatement{
				{
					pos:    tokens.Position{Line: 1, Column: 15},
					Source: "test",
				},
			},
			Items: []ContextItem{
				{
					ContextObject{
						pos:       tokens.Position{Line: 1, Column: 15},
						Private:   false,
						Interface: "test",
						Name:      "bar",
						Extends:   nil,
						Fields:    []FieldStatement{},
						Comment:   "",
					},
				},
				{
					ContextObjectMethod{
						pos:    tokens.Position{Line: 1, Column: 25},
						Target: "foo",
						Name:   "bar",
						Block: FunctionBlock{
							Parameters: FunctionParameters{
								Arguments: ArgumentList{
									pos:   tokens.Position{Line: 1, Column: 35},
									Items: make([]Node, 0),
								},
								ReturnType: nil,
							},
							Body: Block{
								pos:        tokens.Position{Line: 1, Column: 38},
								Statements: []BlockStatement{},
							},
						},
					},
				},
				{
					ContextMethod{
						pos:       tokens.Position{Line: 1, Column: 41},
						Private:   false,
						Interface: "foo",
						Name:      "bar",
						Block: FunctionBlock{
							Parameters: FunctionParameters{
								pos: tokens.Position{Line: 1, Column: 51},
								Arguments: ArgumentList{
									pos:   tokens.Position{Line: 1, Column: 51},
									Items: make([]Node, 0),
								},
								ReturnType: nil,
							},
							Body: Block{
								pos:        tokens.Position{Line: 1, Column: 54},
								Statements: []BlockStatement{},
							},
						},
						Comment: "comment",
					},
				},
				{
					FunctionExpression{
						pos:  tokens.Position{Line: 1, Column: 50},
						Name: "baz",
						Body: FunctionBlock{
							Parameters: FunctionParameters{
								pos: tokens.Position{Line: 1, Column: 51},
								Arguments: ArgumentList{
									pos:   tokens.Position{Line: 1, Column: 51},
									Items: make([]Node, 0),
								},
								ReturnType: nil,
							},
							Body: Block{
								pos:        tokens.Position{Line: 1, Column: 54},
								Statements: []BlockStatement{},
							},
						},
					},
				},
			},
			Comment: "",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE CONTEXT WITH NO OBJECTS
func TestContextNoObjects(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "context foo { }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContext(p)
		},
		expects: &Context{
			pos:     tokens.Position{Line: 1, Column: 1},
			Name:    "foo",
			Remotes: make([]UseStatement, 0),
			Items:   make([]ContextItem, 0),
			Comment: "",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE CONTEXT WITH LEADING COMMENT
func TestContextWithComment(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "//comment\ncontext foo { }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContext(p)
		},
		expects: &Context{
			pos:     tokens.Position{Line: 1, Column: 1},
			Name:    "foo",
			Remotes: make([]UseStatement, 0),
			Items:   make([]ContextItem, 0),
			Comment: "comment",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// ContextObject
// CAN CREATE CONTEXT OBJECT
func TestContextObject(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "test bar { }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextObject(p)
		},
		expects: &ContextObject{
			pos:       tokens.Position{Line: 1, Column: 1},
			Private:   false,
			Interface: "test",
			Name:      "bar",
			Extends:   nil,
			Fields:    []FieldStatement{},
			Comment:   "",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE CONTEXT OBJECT WITH ALL FIELD STATEMENTS
func TestContextObjectWithAllStatements(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "test bar { one = \"abc\" two \"def\" three ghi }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextObject(p)
		},
		expects: &ContextObject{
			pos:       tokens.Position{Line: 1, Column: 1},
			Private:   false,
			Interface: "test",
			Name:      "bar",
			Extends:   nil,
			Fields: []FieldStatement{
				{
					pos: tokens.Position{Line: 1, Column: 10},
					Init: FieldAssignmentExpression{
						pos:  tokens.Position{Line: 1, Column: 10},
						Name: "one",
						Init: Expression{
							pos: tokens.Position{Line: 1, Column: 14},
							Init: Literal{
								pos:   tokens.Position{Line: 1, Column: 14},
								Value: "abc",
							},
						},
					},
				},
				{
					pos: tokens.Position{Line: 1, Column: 20},
					Init: EnumExpression{
						pos:  tokens.Position{Line: 1, Column: 20},
						Name: "two",
						Init: "def",
					},
				},
				{
					pos: tokens.Position{Line: 1, Column: 27},
					Init: FieldExpression{
						pos:  tokens.Position{Line: 1, Column: 27},
						Name: "three",
						Init: TypeExpression{
							pos:        tokens.Position{Line: 1, Column: 34},
							IsArray:    false,
							IsPartial:  false,
							IsOptional: false,
							Selector: Selector{
								pos:     tokens.Position{Line: 1, Column: 34},
								Members: []string{"ghi"},
							},
						},
					},
				},
			},
			Comment: "",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE CONTEXT OBJECT WITH LEADING COMMENT
func TestContextObjectWithComment(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "/*comment*/ test bar { }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextObject(p)
		},
		expects: &ContextObject{
			pos:       tokens.Position{Line: 2, Column: 1},
			Private:   false,
			Interface: "test",
			Name:      "bar",
			Extends:   nil,
			Fields:    []FieldStatement{},
			Comment:   "comment",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE PRIVATE CONTEXT OBJECT
func TestContextObjectPrivate(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "private test bar { }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextObject(p)
		},
		expects: &ContextObject{
			pos:       tokens.Position{Line: 1, Column: 1},
			Private:   true,
			Interface: "test",
			Name:      "bar",
			Extends:   nil,
			Fields:    []FieldStatement{},
			Comment:   "",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE CONTEXT OBJECT WITH EXTENDS
func TestContextObjectExtends(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "test foo extends bar.baz { }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextObject(p)
		},
		expects: &ContextObject{
			pos:       tokens.Position{Line: 1, Column: 1},
			Private:   false,
			Interface: "test",
			Name:      "foo",
			Extends: &Selector{
				pos:     tokens.Position{Line: 1, Column: 16},
				Members: []string{"bar", "baz"},
			},
			Fields:  []FieldStatement{},
			Comment: "",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// ContextObjectMethod
// CAN CREATE CONTEXT OBJECT METHOD
func TestContextObjectMethod(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "func (foo) bar() {}",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextObjectMethod(p)
		},
		expects: &ContextObjectMethod{
			pos:    tokens.Position{Line: 1, Column: 1},
			Target: "foo",
			Name:   "bar",
			Block: FunctionBlock{
				Parameters: FunctionParameters{
					pos: tokens.Position{Line: 1, Column: 13},
					Arguments: ArgumentList{
						pos:   tokens.Position{Line: 1, Column: 13},
						Items: make([]Node, 0),
					},
					ReturnType: nil,
				},
				Body: Block{
					pos:        tokens.Position{Line: 1, Column: 18},
					Statements: []BlockStatement{},
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

// REJECTS EMPTY TARGET
func TestContextObjectMethodEmptyTarget(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "func () bar() {}",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextObjectMethod(p)
		},
		expects:      nil,
		expectsError: ExpectedError(tokens.Position{Line: 1, Column: 7}, tokens.IDENT, ")"),
	})
	if err != nil {
		t.Error(err)
	}
}

// ContextMethod
// CAN CREATE CONTEXT METHOD
func TestContextMethod(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "foo bar() {}",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextMethod(p)
		},
		expects: &ContextMethod{
			pos:       tokens.Position{Line: 1, Column: 1},
			Private:   false,
			Interface: "foo",
			Name:      "bar",
			Block: FunctionBlock{
				Parameters: FunctionParameters{
					pos: tokens.Position{Line: 1, Column: 6},
					Arguments: ArgumentList{
						pos:   tokens.Position{Line: 1, Column: 6},
						Items: make([]Node, 0),
					},
					ReturnType: nil,
				},
				Body: Block{
					pos:        tokens.Position{Line: 1, Column: 11},
					Statements: []BlockStatement{},
				},
			},
			Comment: "",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE PRIVATE CONTEXT METHOD
func TestContextMethodPrivate(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "private foo bar() {}",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextMethod(p)
		},
		expects: &ContextMethod{
			pos:       tokens.Position{Line: 1, Column: 1},
			Private:   true,
			Interface: "foo",
			Name:      "bar",
			Block: FunctionBlock{
				Parameters: FunctionParameters{
					pos: tokens.Position{Line: 1, Column: 11},
					Arguments: ArgumentList{
						pos:   tokens.Position{Line: 1, Column: 11},
						Items: make([]Node, 0),
					},
					ReturnType: nil,
				},
				Body: Block{
					pos:        tokens.Position{Line: 1, Column: 16},
					Statements: []BlockStatement{},
				},
			},
			Comment: "",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE CONTEXT METHOD WITH LEADING COMMENT
func TestContextMethodComment(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "/*comment*/ foo bar() { }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextMethod(p)
		},
		expects: &ContextMethod{
			pos:       tokens.Position{Line: 1, Column: 1},
			Private:   false,
			Interface: "foo",
			Name:      "bar",
			Block: FunctionBlock{
				Parameters: FunctionParameters{
					pos: tokens.Position{Line: 1, Column: 6},
					Arguments: ArgumentList{
						pos:   tokens.Position{Line: 1, Column: 6},
						Items: make([]Node, 0),
					},
					ReturnType: nil,
				},
				Body: Block{
					pos:        tokens.Position{Line: 1, Column: 11},
					Statements: []BlockStatement{},
				},
			},
			Comment: "comment",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// CAN CREATE PRIVATE CONTEXT METHOD WITH LEADING COMMENT
func TestContextMethodPrivateComment(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: "/*comment*/ private foo bar() { }",
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseContextMethod(p)
		},
		expects: &ContextMethod{
			pos:       tokens.Position{Line: 1, Column: 1},
			Private:   true,
			Interface: "foo",
			Name:      "bar",
			Block: FunctionBlock{
				Parameters: FunctionParameters{
					pos: tokens.Position{Line: 1, Column: 16},
					Arguments: ArgumentList{
						pos:   tokens.Position{Line: 1, Column: 16},
						Items: make([]Node, 0),
					},
					ReturnType: nil,
				},
				Body: Block{
					pos:        tokens.Position{Line: 1, Column: 21},
					Statements: []BlockStatement{},
				},
			},
			Comment: "comment",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}
