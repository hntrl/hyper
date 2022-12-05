package nodes

import (
	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"
)

// Manifest :: ImportStatement* Context
type Manifest struct {
	Imports []ImportStatement
	Context Context
}

func (m Manifest) Validate() error {
	for _, imp := range m.Imports {
		if err := imp.Validate(); err != nil {
			return err
		}
	}
	if err := m.Context.Validate(); err != nil {
		return err
	}
	return nil
}

func (m Manifest) Pos() tokens.Position {
	return tokens.Position{Line: 0, Column: 0}
}

func ParseManifest(p *parser.Parser) (*Manifest, error) {
	manifest := Manifest{Imports: make([]ImportStatement, 0)}
	for {
		_, tok, _ := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
		p.Unscan()
		if tok == tokens.IMPORT {
			imp, err := ParseImportStatement(p)
			if err != nil {
				return nil, err
			}
			manifest.Imports = append(manifest.Imports, *imp)
			continue
		}
		break
	}

	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok == tokens.CONTEXT {
		p.Rollback(p.Index() - 2)
		_, tok, _ = p.Scan()
		if tok != tokens.COMMENT {
			p.ScanIgnore(tokens.NEWLINE)
		}
		p.Unscan()
		context, err := ParseContext(p)
		if err != nil {
			return nil, err
		}
		manifest.Context = *context
	} else {
		return nil, ExpectedError(pos, tokens.CONTEXT, lit)
	}
	return &manifest, nil
}

// IMPORT STRING
type ImportStatement struct {
	pos     tokens.Position
	Package string
}

func (i ImportStatement) Validate() error {
	return nil
}

func (i ImportStatement) Pos() tokens.Position {
	return i.pos
}

func ParseImportStatement(p *parser.Parser) (*ImportStatement, error) {
	pos, tok, lit := p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.IMPORT {
		return nil, ExpectedError(pos, tokens.IMPORT, lit)
	}
	stmt := ImportStatement{pos: pos}
	pos, tok, lit = p.ScanIgnore(tokens.NEWLINE, tokens.COMMENT)
	if tok != tokens.STRING {
		return nil, ExpectedError(pos, tokens.STRING, lit)
	}
	stmt.Package = lit
	return &stmt, nil
}
