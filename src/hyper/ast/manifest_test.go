package ast

import (
	"testing"

	"github.com/hntrl/hyper/src/hyper/parser"
	"github.com/hntrl/hyper/src/hyper/tokens"
)

// Manifest
// CAN PARSE MANIFEST
func TestManifest(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `import "foo" context bar {}`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseManifest(p)
		},
		expects: &Manifest{
			Imports: []ImportStatement{
				{
					pos:    tokens.Position{Line: 1, Column: 1},
					Source: "foo",
				},
			},
			Context: Context{
				pos:     tokens.Position{Line: 1, Column: 14},
				Name:    "bar",
				Remotes: []UseStatement{},
				Items:   []ContextItem{},
				Comment: "",
			},
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}

// ImportStatement
// CAN PARSE IMPORT STATEMENT
func TestImportStatement(t *testing.T) {
	err := evaluateTest(TestFixture{
		lit: `import "foo"`,
		parseFn: func(p *parser.Parser) (Node, error) {
			return ParseImportStatement(p)
		},
		expects: &ImportStatement{
			pos:    tokens.Position{Line: 1, Column: 1},
			Source: "foo",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}
