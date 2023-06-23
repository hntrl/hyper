package parser

import (
	"bufio"
	"fmt"
	"io"
	"unicode"

	"github.com/hntrl/hyper/internal/tokens"
)

type LexerError struct {
	pos tokens.Position
	err error
}

func (le LexerError) Error() string {
	return fmt.Sprintf("(%s) %s", le.pos.String(), le.err.Error())
}

type Lexer struct {
	pos    tokens.Position
	reader *bufio.Reader
}

func NewLexer(reader *bufio.Reader) *Lexer {
	return &Lexer{
		pos:    tokens.Position{Line: 1, Column: 0},
		reader: reader,
	}
}

func (l *Lexer) read() rune {
	r, _, err := l.reader.ReadRune()
	if err != nil {
		panic(LexerError{l.pos, err})
	}
	l.pos.Column++
	return r
}
func (l *Lexer) backup() {
	l.pos.Column--
	l.reader.UnreadRune()
}
func (l *Lexer) peek() rune {
	r := l.read()
	l.backup()
	return r
}
func (l *Lexer) resetPosition() {
	l.pos.Line++
	l.pos.Column = 0
}

func (l *Lexer) digits() string {
	var lit string
	for {
		r := l.read()
		if r == '_' {
			continue
		}
		if unicode.IsDigit(r) {
			lit += string(r)
			continue
		}
		l.backup()
		break
	}
	return lit
}

func (l *Lexer) lexIdent() string {
	var lit string
	for {
		r := l.read()
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			lit += string(r)
		} else {
			l.backup()
			return lit
		}
	}
}
func (l *Lexer) lexNumber() (tokens.Token, string) {
	r := l.read()
	if r == '.' {
		return tokens.FLOAT, "." + l.digits()
	} else {
		l.backup()
		first := l.digits()
		if l.peek() == '.' {
			l.read()
			second := l.digits()
			return tokens.FLOAT, first + "." + second
		} else {
			return tokens.INT, first
		}
	}
}
func (l *Lexer) lexEscaped() string {
	var lit string
	r := l.read()
	switch r {
	case '\\':
		lit += "\\"
	case 'n':
		lit += "\n"
	case 't':
		lit += "\t"
	case 'r':
		lit += "\r"
	case 'x':
		lit += "\\x" + l.digits()
	case '\'':
		lit += "'"
	case '"':
		lit += "\""
	default:
		lit += string(r)
	}
	return lit
}
func (l *Lexer) lexString() string {
	var lit string

	terminator := l.read()
	if terminator != '"' && terminator != '\'' {
		l.backup()
		return ""
	}
	for {
		r := l.read()
		switch r {
		case '\\':
			l.backup()
			lit += l.lexEscaped()
		case terminator:
			return lit
		default:
			lit += string(r)
		}
	}
}
func (l *Lexer) lexComment() string {
	var lit string
	first := l.read()

	if first == '/' {
		for {
			r := l.read()
			if r == '\n' {
				l.resetPosition()
				for {
					// i have no idea why this chain works BUT IT DOES!
					r := l.read()
					if r == '\n' {
						l.backup()
						break
					} else if unicode.IsSpace(r) {
						continue
					} else if r == '/' {
						lit += "\\n" + l.lexComment()
					} else {
						l.backup()
					}
					break
				}
				break
			} else if r == 0 {
				break
			} else {
				lit += string(r)
			}
		}
	} else if first == '*' {
		for {
			r := l.read()
			if r == '\n' {
				l.resetPosition()
				lit += string(r)
			} else if r == '*' {
				if l.peek() == '/' {
					l.read()
					break
				}
			} else if r == 0 {
				break
			} else {
				lit += string(r)
			}
		}
	}
	return lit
}

func (l *Lexer) Lex() (tokens.Position, tokens.Token, string) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return l.pos, tokens.EOF, ""
			}
			panic(LexerError{l.pos, err})
		}
		l.pos.Column++
		if unicode.IsSpace(r) && r != '\n' {
			continue
		}

		var tok tokens.Token
		var lit string
		startPos := l.pos

		if unicode.IsLetter(r) {
			l.backup()
			lit = l.lexIdent()
			tok = tokens.Lookup(lit)
		} else if unicode.IsDigit(r) || r == '.' && unicode.IsDigit(l.peek()) {
			l.backup()
			tok, lit = l.lexNumber()
		} else {
			switch r {
			case '\n':
				l.resetPosition()
				tok, lit = tokens.NEWLINE, "\\n"
			case '"', '\'':
				l.backup()
				tok, lit = tokens.STRING, l.lexString()
			case '&':
				if l.peek() == '&' {
					l.read()
					tok, lit = tokens.AND, "&&"
				} else {
					tok, lit = tokens.ILLEGAL, "&"
				}
			case '|':
				if l.peek() == '|' {
					l.read()
					tok, lit = tokens.OR, "||"
				} else {
					tok, lit = tokens.ILLEGAL, "|"
				}
			case '+':
				newToken := l.assignDupeSingleSwitch(tokens.ADD, tokens.ADD_ASSIGN, '+', tokens.INC)
				tok, lit = newToken, newToken.String()
			case '-':
				newToken := l.assignDupeSingleSwitch(tokens.SUB, tokens.SUB_ASSIGN, '-', tokens.DEC)
				tok, lit = newToken, newToken.String()
			case '*':
				newToken := l.assignDupeSwitch(tokens.MUL, tokens.MUL_ASSIGN, '*', tokens.PWR, tokens.PWR_ASSIGN)
				tok, lit = newToken, newToken.String()
			case '/':
				if l.peek() == '/' || l.peek() == '*' {
					tok, lit = tokens.COMMENT, l.lexComment()
				} else {
					newToken := l.assignSwitch(tokens.QUO, tokens.QUO_ASSIGN)
					tok, lit = newToken, newToken.String()
				}
			case '%':
				newToken := l.assignSwitch(tokens.REM, tokens.REM_ASSIGN)
				tok, lit = newToken, newToken.String()
			case '=':
				newToken := l.assignSwitch(tokens.ASSIGN, tokens.EQUALS)
				tok, lit = newToken, newToken.String()
			case '<':
				newToken := l.assignSwitch(tokens.LESS, tokens.LESS_EQUAL)
				tok, lit = newToken, newToken.String()
			case '>':
				newToken := l.assignSwitch(tokens.GREATER, tokens.GREATER_EQUAL)
				tok, lit = newToken, newToken.String()
			case '!':
				newToken := l.assignSwitch(tokens.NOT, tokens.NOT_EQUALS)
				tok, lit = newToken, newToken.String()
			case '.':
				if l.peek() == '.' {
					l.read()
					l.read()
					tok, lit = tokens.ELLIPSIS, "..."
				} else {
					tok, lit = tokens.PERIOD, "."
				}
			case ',':
				tok, lit = tokens.COMMA, ","
			case ';':
				tok, lit = tokens.SEMICOLON, ";"
			case ':':
				newToken := l.assignSwitch(tokens.COLON, tokens.DEFINE)
				tok, lit = newToken, newToken.String()
			case '?':
				tok, lit = tokens.QUESTION, "?"
			case '{':
				tok, lit = tokens.LCURLY, "{"
			case '}':
				tok, lit = tokens.RCURLY, "}"
			case '[':
				tok, lit = tokens.LSQUARE, "["
			case ']':
				tok, lit = tokens.RSQUARE, "]"
			case '(':
				tok, lit = tokens.LPAREN, "("
			case ')':
				tok, lit = tokens.RPAREN, ")"
			case '`':
				tok, lit = tokens.BACKTICK, "`"
			default:
				tok, lit = tokens.ILLEGAL, string(r)
			}
		}
		return startPos, tok, lit
	}
}

func (l *Lexer) assignSwitch(tok0, tok1 tokens.Token) tokens.Token {
	if l.peek() == '=' {
		l.read()
		return tok1
	}
	return tok0
}

func (l *Lexer) assignDupeSingleSwitch(tok0, tok1 tokens.Token, ch2 rune, tok2 tokens.Token) tokens.Token {
	if l.peek() == '=' {
		l.read()
		return tok1
	}
	if l.peek() == ch2 {
		l.read()
		return tok2
	} else {
		return tok0
	}
}

func (l *Lexer) assignDupeSwitch(tok0, tok1 tokens.Token, ch2 rune, tok2 tokens.Token, tok3 tokens.Token) tokens.Token {
	if l.peek() == '=' {
		l.read()
		return tok1
	}
	if l.peek() == ch2 {
		l.read()
		if l.peek() == '=' {
			l.read()
			return tok3
		}
		return tok2
	}
	return tok0
}
