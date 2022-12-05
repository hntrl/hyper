package nodes

import (
	"testing"

	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"
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
					pos:     tokens.Position{Line: 1, Column: 1},
					Package: "foo",
				},
			},
			Context: Context{
				pos:     tokens.Position{Line: 1, Column: 14},
				Name:    "bar",
				Objects: []Node{},
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
			pos:     tokens.Position{Line: 1, Column: 1},
			Package: "foo",
		},
		expectsError: nil,
		endingToken:  tokens.EOF,
	})
	if err != nil {
		t.Error(err)
	}
}
