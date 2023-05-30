## Grammar & AST Props

```
Manifest :: ImportStatement* Context
  * Imports: []ImportStatement
  * Context: Context

ImportStatement :: IMPORT STRING
  * Package: string

Selector :: IDENT
          | IDENT PERIOD Selector
  * Members: []string

Literal :: STRING
         | INT
         | FLOAT
  * Value: interface{}

Context :: COMMENT? CONTEXT Selector LCURLY (UseStatement | ContextItem)* RCURLY
  * Name: string
  * Remotes: []UseStatement
  * Items: []ContextItem
  * Comment: string

ContextItemSet :: ContextItem*
  * Items: []ContextItem

ContextItem :: (ContextObject | ContextObjectMethod | ContextMethod | RemoteContextMethod | FunctionExpression)
  * Init: (ContextObject | ContextObjectMethod | ContextMethod | RemoteContextMethod | FunctionExpression)

UseStatement :: USE IDENT
  * Source: string

ContextObject :: COMMENT? PRIVATE? IDENT IDENT (EXTENDS Selector)? LCURLY FieldStatement* RCURLY
  * Private: bool
  * Interface: string
  * Name: string
  * Extends: Selector
  * Fields: []FieldStatement
  * Comment: string

ContextObjectMethod :: FUNC LPAREN IDENT RPAREN IDENT FunctionBlock
  * Target: string
  * Name: string
  * Block: FunctionBlock

ContextMethod :: COMMENT? PRIVATE? IDENT IDENT FunctionBlock
  * Private: bool
  * Interface: string
  * Name: string
  * Block: FunctionBlock
  * Comment: string

RemoteContextMethod :: COMMENT? PRIVATE? REMOTE IDENT IDENT FunctionParameters
  * Private: bool
  * Interface: string
  * Name: string
  * Parameters: FunctionParameters
  * Comment: string

FieldStatement :: COMMENT? AssignmentStatement
                | COMMENT? EnumStatement
                | COMMENT? TypeStatement
  * Init: (AssignmentStatement | EnumStatement | TypeStatement)
  * Comment: string

AssignmentStatement :: IDENT ASSIGN Expression
  * Name: string
  * Init: Expression

EnumStatement :: IDENT STRING
  * Name: string
  * Value: string

TypeStatement :: IDENT TypeExpression
  * Name: string
  * Init: TypeExpression

FunctionStatement :: IDENT FunctionBlock
  * Name: string
  * Init: FunctionBlock

TypeExpression :: (LSQUARE RSQUARE)? Selector QUESTION?
                | (LSQUARE RSQUARE) PARTIAL LT Selector GT QUESTION?
  * IsArray: bool
  * IsPartial: bool
  * IsOptional: bool
  * Members: []string

Expression :: Literal
            | ArrayExpression
            | InstanceExpression
            | UnaryExpression
            | BinaryExpression
            | ObjectPattern
            | FunctionExpression
            | ValueExpression
            | LPAREN Expression RPAREN
  * Init: (Literal | ArrayExpression | InstanceExpression | UnaryExpression | BinaryExpression | ObjectPattern | FunctionExpression | ValueExpression | Expression)

ArrayExpression :: LSQUARE RSQUARE TypeExpression LCURLY ((Expression COMMA) | Expression)* RCURLY
  * Init: TypeExpression
  * Elements: []Expression

InstanceExpression :: Selector LCURLY PropertyList RCURLY
  * Selector: Selector
  * Properties: PropertyList

UnaryExpression :: NOT Expression
                 | ADD Expression
                 | SUB Expression
  * Operator: token(NOT | ADD | SUB)
  * Init: Expression

BinaryExpression :: Expression token(IsOperator) Expression
  * Left: Expression
  * Operator: token(IsOperator)
  * Right: Expression

ValueExpression :: IDENT ValueExpressionMember*
  * Members: []ValueExpressionMember

ValueExpressionMember :: PERIOD IDENT
                       | CallExpression
                       | IndexExpression
  * Init: (string | CallExpression | IndexExpression)

CallExpression :: LPAREN ((Expression COMMA) | Expression)* RPAREN
  * Arguments: []Expression

IndexExpression :: LSQUARE Expression? SEMICOLON? Expression? RSQUARE
  * Left: Expression?
  * IsRange: bool
  * Right: Expression?

AssignmentExpression :: Selector token(IsAssignmentOperator) Expression
                      | Selector (INC | DEC)
  * Name: Selector
  * Operator: token(IsAssignmentOperator) | (INC | DEC)
  * Init: Expression

ObjectPattern :: LCURLY PropertyList RCURLY
  * Properties: PropertyList

PropertyList :: (Property | SpreadElement) (COMMA PropertyList)?
  = [](Property | SpreadElement)

Property :: IDENT COLON Expression
  * Key: string
  * Init: Expression

SpreadElement :: ELLIPSIS Expression
  * Init: Expression

ArgumentList :: (ArgumentItem | ArgumentObject) (COMMA ArgumentList)?
  * Items: [](ArgumentItem | ArgumentObject)

ArgumentItem :: IDENT COLON TypeExpression
  * Key: string
  * Init: TypeExpression

ArgumentObject :: LCURLY (ArgumentItem)? (ArgumentItem COMMA)* RCURLY
  * Items: []ArgumentItem

FunctionParameters :: LPAREN ArgumentList? RPAREN TypeExpression?
  * Arguments: ArgumentList
  * ReturnType: TypeExpression?

FunctionBlock :: FunctionParameters LCURLY Block RCURLY
  * Parameters: FunctionParameters
  * ReturnType: TypeExpression?
  * Body: Block

FunctionExpression :: FUNC IDENT FunctionBlock
  * Name: string
  * Body: Block

Block :: BlockStatement*
  * Statements: []BlockStatement

InlineBlock :: (BlockStatement | LCURLY Block RCURLY)
  * Body: Block

BlockStatement :: Expression
                | DeclarationStatement
                | AssignmentExpression
                | IfStatement
                | WhileStatement
                | ForStatement
                | ContinueStatement
                | BreakStatement
                | SwitchBlock
                | GuardStatement
                | ReturnStatement
                | ThrowStatement
  * Init: (Expression | DeclarationStatement | AssignmentExpression | IfStatement | WhileStatement | ForStatement | ContinueStatement | BreakStatement | SwitchBlock | GuardStatement | ReturnStatement | ThrowStatement)

DeclarationStatement :: IDENT DEFINE Expression
  * Name: string
  * Init: Expression

IfStatement :: IF LPAREN Expression RPAREN InlineBlock (ELSE IfStatement)? (ELSE Block)?
  * Condition: Expression
  * Body: Block
  * Alternate: (IfStatement | Block)?

WhileStatement :: WHILE LPAREN Expression RPAREN InlineBlock
  * Condition: Expression
  * Body: Block

ForStatement :: FOR LPAREN (ForCondition | RangeCondition) RPAREN InlineBlock
  * Condition: (ForCondition | RangeCondition)
  * Body: Block

ForCondition :: (DeclarationStatement | Expression) SEMICOLON Expression (SEMICOLON (Expression | AssignmentExpression))?
  * Init: DeclarationStatement?
  * Condition: Expression
  * Update: (Expression | AssignmentExpression)

RangeCondition :: IDENT COMMA IDENT IN Expression
  * Index: string
  * Value: string
  * Target: Expression

ContinueStatement :: CONTINUE
  = nil

BreakStatement :: BREAK
  = nil

SwitchBlock :: SWITCH LPAREN Expression RPAREN LCURLY SwitchStatement* RCURLY
  * Target: Expression
  * Statements: []SwitchStatement

SwitchStatement :: ((CASE Expression) | DEFAULT) COLON Block
  * Condition: Expression
  * IsDefault: bool
  * Body: Block

GuardStatement :: GUARD Expression
  * Init: Expression

ReturnStatement :: RETURN Expression
  * Init: Expression

ThrowStatement :: THROW Expression
  * Init: Expression
```
