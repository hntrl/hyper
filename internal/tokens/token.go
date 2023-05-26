package tokens

type Token int

const (
	EOF Token = iota
	ILLEGAL
	COMMENT
	NEWLINE

	// Identifiers and type literals
	IDENT  // context
	INT    // 215
	FLOAT  // 215.34
	STRING // "abc"

	operator_beg
	// Operators and delimiters
	ADD // +
	SUB // -
	MUL // *
	PWR // **
	QUO // /
	REM // %
	operator_end

	comparator_beg
	AND // &&
	OR  // ||

	EQUALS  // ==
	LESS    // <
	GREATER // >
	NOT     // !

	NOT_EQUALS    // !=
	LESS_EQUAL    // <=
	GREATER_EQUAL // >=
	comparator_end

	assign_beg
	// Assignment operators
	ASSIGN     // =
	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	PWR_ASSIGN // **=
	QUO_ASSIGN // /=
	REM_ASSIGN // %=
	INC        // ++
	DEC        // --
	assign_end

	DEFINE   // :=
	ELLIPSIS // ...

	PERIOD    // .
	COMMA     // ,
	SEMICOLON // ;
	COLON     // :
	QUESTION  // ?

	LCURLY  // {
	RCURLY  // }
	LSQUARE // [
	RSQUARE // ]
	LPAREN  // (
	RPAREN  // )

	keyword_beg
	// Keywords
	IMPORT
	CONTEXT
	USE
	PRIVATE
	REMOTE
	EXTENDS
	FUNC
	VAR
	IF
	ELSE
	WHILE
	FOR
	IN
	CONTINUE
	BREAK
	SWITCH
	CASE
	DEFAULT
	GUARD
	RETURN
	THROW
	PARTIAL
	keyword_end
)

var tokens = []string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",
	COMMENT: "COMMENT",
	NEWLINE: "NEWLINE",

	IDENT:  "identifier",
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",

	AND: "&&",
	OR:  "||",
	INC: "++",
	DEC: "--",

	ADD_ASSIGN: "+=",
	SUB_ASSIGN: "-=",
	MUL_ASSIGN: "*=",
	PWR_ASSIGN: "**=",
	QUO_ASSIGN: "/=",
	REM_ASSIGN: "%=",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	PWR: "**",
	QUO: "/",
	REM: "%",

	EQUALS:  "==",
	LESS:    "<",
	GREATER: ">",
	NOT:     "!",

	ASSIGN:        "=",
	NOT_EQUALS:    "!=",
	LESS_EQUAL:    "<=",
	GREATER_EQUAL: ">=",
	DEFINE:        ":=",
	ELLIPSIS:      "...",

	PERIOD:    ".",
	COMMA:     ",",
	SEMICOLON: ";",
	COLON:     ":",
	QUESTION:  "?",

	LCURLY:  "{",
	RCURLY:  "}",
	LSQUARE: "[",
	RSQUARE: "]",
	LPAREN:  "(",
	RPAREN:  ")",

	IMPORT:   "import",
	CONTEXT:  "context",
	USE:      "use",
	PRIVATE:  "private",
	REMOTE:   "remote",
	EXTENDS:  "extends",
	FUNC:     "func",
	VAR:      "var",
	IF:       "if",
	ELSE:     "else",
	WHILE:    "while",
	FOR:      "for",
	IN:       "in",
	CONTINUE: "continue",
	BREAK:    "break",
	SWITCH:   "switch",
	CASE:     "case",
	DEFAULT:  "default",
	GUARD:    "guard",
	RETURN:   "return",
	THROW:    "throw",
	PARTIAL:  "Partial",
}

const (
	LowestPrec = 0
	UnaryPrec  = 7
)

func (op Token) Precedence() int {
	switch op {
	case OR:
		return 1
	case AND:
		return 2
	case EQUALS, NOT_EQUALS, LESS, LESS_EQUAL, GREATER, GREATER_EQUAL:
		return 3
	case ADD, SUB:
		return 4
	case MUL, QUO, REM:
		return 5
	case PWR:
		return 6
	}
	return LowestPrec
}

func (t Token) String() string {
	return tokens[t]
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token, keyword_end-(keyword_beg+1))
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

func Lookup(ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return IDENT
}

func (t Token) IsKeyword() bool {
	return keyword_beg < t && t < keyword_end
}

func (t Token) IsOperator() bool {
	return operator_beg < t && t < operator_end
}

func (t Token) IsComparableOperator() bool {
	return comparator_beg < t && t < comparator_end
}

func (t Token) IsAssignmentOperator() bool {
	return assign_beg < t && t < assign_end
}
