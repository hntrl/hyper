package language

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hntrl/lang/language/nodes"
	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"
)

func ParseFromFile(path string) (*nodes.Manifest, error) {
	errorHandler := func(pos tokens.Position, msg string) {
		panic(fmt.Sprintf("%s, %s", pos.String(), msg))
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	lexer := parser.NewLexer(bufio.NewReader(file), errorHandler)
	parser := parser.NewParser(lexer)

	manifest, err := nodes.ParseManifest(parser)
	if err != nil {
		return nil, err
	}
	err = manifest.Validate()
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

// func ParseFromFile(path string) (*nodes.Manifest, error) {
// 	errorHandler := func(pos tokens.Position, msg string) {
// 		panic(fmt.Sprintf("%s, %s", pos.String(), msg))
// 	}

// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	lexer := NewLexer(bufio.NewReader(file), errorHandler)

// 	for {
// 		pos, tok, lit := lexer.Lex()
// 		if tok == tokens.EOF {
// 			break
// 		}

// 		fmt.Printf("%d:%d\t%s\t%s\n", pos.Line, pos.Column, tok, lit)
// 	}

// 	return nil, nil
// }
