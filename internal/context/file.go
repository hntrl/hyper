package context

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/parser"
	"github.com/hntrl/hyper/internal/tokens"
)

// FIXME: locate these methods into somewhere other than context/

func ParseContextFromFile(path string) (*ast.Manifest, error) {
	errorHandler := func(pos tokens.Position, msg string) {
		panic(fmt.Sprintf("%s, %s", pos.String(), msg))
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	lexer := parser.NewLexer(bufio.NewReader(file), errorHandler)
	parser := parser.NewParser(lexer)

	manifest, err := ast.ParseManifest(parser)
	if err != nil {
		return nil, err
	}
	err = manifest.Validate()
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func ParseContextItemSetFromFile(path string) (*ast.ContextItemSet, error) {
	errorHandler := func(pos tokens.Position, msg string) {
		panic(fmt.Sprintf("%s, %s, %s", path, pos.String(), msg))
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	lexer := parser.NewLexer(bufio.NewReader(file), errorHandler)
	parser := parser.NewParser(lexer)

	items, err := ast.ParseContextItemSet(parser)
	if err != nil {
		return nil, err
	}
	err = items.Validate()
	if err != nil {
		return nil, err
	}
	return items, nil
}
