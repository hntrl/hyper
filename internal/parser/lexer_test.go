package parser

import (
	"bufio"
	"strings"
	"testing"

	"github.com/hntrl/hyper/internal/tokens"
)

func setupLexer(lit string) Lexer {
	reader := bufio.NewReader(strings.NewReader(lit))
	return *NewLexer(reader)
}

// lextIdent()

// CAN IDENTIFY IDENT TOKENS
func TestIdent(t *testing.T) {
	// Setup
	lexer := setupLexer("abc")
	lit := lexer.lexIdent()

	// Assert
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
}

// CAN IDENTIFY IDENT TOKENS WITH NUMBERS
func TestIdentWithNumbers(t *testing.T) {
	// Setup
	lexer := setupLexer("abc123")
	lit := lexer.lexIdent()

	// Assert
	if lit != "abc123" {
		t.Errorf("Expected abc123, got %v", lit)
	}
}

// lexNumber()

// CAN PARSE INTEGER
func TestParseInteger(t *testing.T) {
	// Setup
	lexer := setupLexer("123")
	tok, lit := lexer.lexNumber()

	// Assert
	if tok != tokens.INT {
		t.Errorf("Expected INT, got %v", tok)
	}
	if lit != "123" {
		t.Errorf("Expected 123, got %v", lit)
	}
}

// CAN PARSE FLOAT (STARTING WITH 0)
func TestParseFloatStartingWithZero(t *testing.T) {
	// Setup
	lexer := setupLexer("0.123")
	tok, lit := lexer.lexNumber()

	// Assert
	if tok != tokens.FLOAT {
		t.Errorf("Expected FLOAT, got %v", tok)
	}
	if lit != "0.123" {
		t.Errorf("Expected 0.123, got %v", lit)
	}
}

// CAN PARSE FLOAT (STARTING WITH .)
func TestParseFloatStartingWithDot(t *testing.T) {
	// Setup
	lexer := setupLexer(".123")
	tok, lit := lexer.lexNumber()

	// Assert
	if tok != tokens.FLOAT {
		t.Errorf("Expected FLOAT, got %v", tok)
	}
	if lit != ".123" {
		t.Errorf("Expected .123, got %v", lit)
	}
}

// DOESN'T INCLUDE MULTIPLE DECIMAL POINTS IN RETURNED LITERAL
func TestParseFloatWithMultipleDecimalPoints(t *testing.T) {
	// Setup
	lexer := setupLexer("0.1.23")
	tok, lit := lexer.lexNumber()

	// Assert
	if tok != tokens.FLOAT {
		t.Errorf("Expected FLOAT, got %v", tok)
	}
	if lit != "0.1" {
		t.Errorf("Expected 0.1, got %v", lit)
	}
}

// lexString()
// CAN IDENTIFY A SINGLE QUOTE STRING
func TestLexSingleQuoteString(t *testing.T) {
	// Setup
	lexer := setupLexer("'abc'")
	lit := lexer.lexString()

	// Assert
	if lit != "abc" {
		t.Errorf("Expected 'abc', got %v", lit)
	}
}

// CAN IDENTIFY A DOUBLE QUOTE STRING
func TestLexDoubleQuoteString(t *testing.T) {
	// Setup
	lexer := setupLexer("\"abc\"")
	lit := lexer.lexString()

	// Assert
	if lit != "abc" {
		t.Errorf("Expected \"abc\", got %v", lit)
	}
}

// CAN HANDLE ESCAPE CHARACTERS
func TestLexStringWithEscapeCharacters(t *testing.T) {
	// Setup
	testString := "a\\\"\\\\b"
	lexer := setupLexer("\"" + testString + "\"")
	lit := lexer.lexString()

	// Assert
	if lit != "a\"\\b" {
		t.Errorf("Expected a\"\\b, got %v", lit)
	}
}

// lexComment()
// CAN IDENTIFY A MULTI-LINE COMMENT
func TestLexMultiLineComment(t *testing.T) {
	// Setup
	lexer := setupLexer("*abc*/")
	lit := lexer.lexComment()

	// Assert
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
}

// CAN IDENTIFY A SINGLE-LINE COMMENT
func TestLexSingleLineComment(t *testing.T) {
	// Setup
	lexer := setupLexer("/abc")
	lit := lexer.lexComment()

	// Assert
	if lit != "abc" {
		t.Errorf("Expected abc, got %v", lit)
	}
}

// CAN GROUP SINGLE LINE COMMENTS TOGETHER INTO A SINGLE BLOCK
func TestLexMultipleSingleLineComments(t *testing.T) {
	// Setup
	lexer := setupLexer("/abc\n//def\n//ghi")
	lit := lexer.lexComment()

	// Assert
	if lit != "abc\\ndef\\nghi" {
		t.Errorf("Expected abc\\ndef\\nghi, got %v", lit)
	}
}

// CAN GROUP SINGLE LINE COMMENTS TOGETHER INTO A SINGLE BLOCK (IGNORING BEGINNING SPACES)
func TestLexMultipleSingleLineCommentsWithSpaces(t *testing.T) {
	// Setup
	lexer := setupLexer("/abc\n //def\n  //ghi")
	lit := lexer.lexComment()

	// Assert
	if lit != "abc\\ndef\\nghi" {
		t.Errorf("Expected abc\\ndef\\nghi, got %v", lit)
	}
}

// CAN SEPARATE COMMENT BLOCKS BY NEWLINE
func TestLexMultipleCommentBlocks(t *testing.T) {
	// Setup
	lexer := setupLexer("/abc\n//def\n\n//ghi")
	lit := lexer.lexComment()

	// Assert
	if lit != "abc\\ndef" {
		t.Errorf("Expected abc\\ndef, got %v", lit)
	}

	lexer.read()
	lexer.read()

	lit = lexer.lexComment()
	if lit != "ghi" {
		t.Errorf("Expected ghi, got %v", lit)
	}
}

// Lexer.Lex()
// CAN TOKENIZE ALL TERMINALS CORRECTLY
func TestLexTerminals(t *testing.T) {
	type TokenFixture struct {
		Token   tokens.Token
		Literal string
	}
	// Setup
	expected := []TokenFixture{
		{tokens.AND, "&&"},
		{tokens.OR, "||"},
		{tokens.ADD, "+"},
		{tokens.SUB, "-"},
		{tokens.MUL, "*"},
		{tokens.PWR, "**"},
		{tokens.QUO, "/"},
		{tokens.REM, "%"},
		{tokens.EQUALS, "=="},
		{tokens.LESS, "<"},
		{tokens.GREATER, ">"},
		{tokens.NOT, "!"},
		{tokens.NOT_EQUALS, "!="},
		{tokens.LESS_EQUAL, "<="},
		{tokens.GREATER_EQUAL, ">="},
		{tokens.ASSIGN, "="},
		{tokens.ADD_ASSIGN, "+="},
		{tokens.SUB_ASSIGN, "-="},
		{tokens.MUL_ASSIGN, "*="},
		{tokens.PWR_ASSIGN, "**="},
		{tokens.QUO_ASSIGN, "/="},
		{tokens.REM_ASSIGN, "%="},
		{tokens.INC, "++"},
		{tokens.DEC, "--"},
		{tokens.DEFINE, ":="},
		{tokens.ELLIPSIS, "..."},
		{tokens.PERIOD, "."},
		{tokens.SEMICOLON, ";"},
		{tokens.COLON, ":"},
		{tokens.QUESTION, "?"},
		{tokens.LCURLY, "{"},
		{tokens.RCURLY, "}"},
		{tokens.LSQUARE, "["},
		{tokens.RSQUARE, "]"},
		{tokens.LPAREN, "("},
		{tokens.RPAREN, ")"},
	}

	strParts := []string{}
	for _, fixture := range expected {
		strParts = append(strParts, fixture.Literal)
	}
	lexer := setupLexer(strings.Join(strParts, " "))

	// Assert
	for _, expected := range expected {
		_, tok, lit := lexer.Lex()
		if tok != expected.Token {
			t.Errorf("Expected %v, got %v", expected.Token, tok)
		}
		if lit != expected.Literal {
			t.Errorf("Expected %v, got %v", expected.Literal, lit)
		}
	}
}

// CAN TOKENIZE ALL KEYWORDS CORRECTLY
func TestLexKeywords(t *testing.T) {
	type TokenFixture struct {
		Token   tokens.Token
		Literal string
	}
	// Setup
	expected := []TokenFixture{
		{tokens.IMPORT, "import"},
		{tokens.CONTEXT, "context"},
		{tokens.PRIVATE, "private"},
		{tokens.EXTENDS, "extends"},
		{tokens.FUNC, "func"},
		{tokens.VAR, "var"},
		{tokens.IF, "if"},
		{tokens.ELSE, "else"},
		{tokens.WHILE, "while"},
		{tokens.FOR, "for"},
		{tokens.IN, "in"},
		{tokens.CONTINUE, "continue"},
		{tokens.BREAK, "break"},
		{tokens.SWITCH, "switch"},
		{tokens.CASE, "case"},
		{tokens.DEFAULT, "default"},
		{tokens.RETURN, "return"},
		{tokens.THROW, "throw"},
		// check this doesn't become a keyword
		{tokens.IDENT, "abc"},
	}

	strParts := []string{}
	for _, fixture := range expected {
		strParts = append(strParts, fixture.Literal)
	}
	lexer := setupLexer(strings.Join(strParts, " "))

	// Assert
	for _, expected := range expected {
		_, tok, lit := lexer.Lex()
		if tok != expected.Token {
			t.Errorf("Expected %v, got %v", expected.Token, tok)
		}
		if lit != expected.Literal {
			t.Errorf("Expected %v, got %v", expected.Literal, lit)
		}
	}
}

// CAN IDENTIFY ILLEGAL TOKENS
func TestLexIllegal(t *testing.T) {
	// Setup
	lexer := setupLexer("$")

	// Assert
	_, tok, lit := lexer.Lex()
	if tok != tokens.ILLEGAL {
		t.Errorf("Expected illegal, got %v", tok)
	}
	if lit != "$" {
		t.Errorf("Expected $, got %v", lit)
	}
}
