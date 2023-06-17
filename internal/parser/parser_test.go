package parser

import (
	"bufio"
	"strings"
	"testing"

	"github.com/hntrl/hyper/internal/tokens"
)

func setupParser(lit string) *Parser {
	reader := bufio.NewReader(strings.NewReader(lit))
	lexer := NewLexer(reader)
	return NewParser(lexer)
}

// CAN SCAN THE NEXT TOKEN
func TestScan(t *testing.T) {
	// Setup
	parser := setupParser("abc def")

	// Assert
	_, tok, lit := parser.Scan()
	if tok != tokens.IDENT {
		t.Errorf("Expected IDENT, got %v", tok)
	}
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
	_, tok, lit = parser.Scan()
	if tok != tokens.IDENT {
		t.Errorf("Expected IDENT, got %v", tok)
	}
	if lit != "def" {
		t.Errorf("Expected def, got %v", lit)
	}
}

// CAN SCAN THE NEXT TOKEN IF ROLLED BACK
func TestScanAfterUnscan(t *testing.T) {
	// Setup
	parser := setupParser("abc def")

	// Assert
	_, tok, lit := parser.Scan()
	if tok != tokens.IDENT {
		t.Errorf("Expected IDENT, got %v", tok)
	}
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
	parser.Unscan()
	_, tok, lit = parser.Scan()
	if tok != tokens.IDENT {
		t.Errorf("Expected IDENT, got %v", tok)
	}
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
}

// CAN SCAN IGNORING TOKENS
func TestScanIgnoring(t *testing.T) {
	// Setup
	parser := setupParser("abc = def")

	// Assert
	_, tok, lit := parser.Scan()
	if tok != tokens.IDENT {
		t.Errorf("Expected IDENT, got %v", tok)
	}
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
	_, tok, lit = parser.ScanIgnore(tokens.ASSIGN)
	if tok != tokens.IDENT {
		t.Errorf("Expected IDENT, got %v", tok)
	}
	if lit != "def" {
		t.Errorf("Expected def, got %v", lit)
	}
}

// CAN ROLLBACK TO AN INDEX
func TestUnscanTo(t *testing.T) {
	// Setup
	parser := setupParser("abc def")

	// Assert
	_, tok, lit := parser.Scan()
	if tok != tokens.IDENT {
		t.Errorf("Expected IDENT, got %v", tok)
	}
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
	parser.Rollback(-1)
	_, tok, lit = parser.Scan()
	if tok != tokens.IDENT {
		t.Errorf("Expected IDENT, got %v", tok)
	}
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
}
