## Grammar & AST Props

```
Manifest :: ImportStatement* Context
  * Imports: []ImportStatement
  * Context: Context

ImportStatement :: IMPORT STRING
  * Source: string

Selector :: IDENT
          | IDENT PERIOD Selector
  * Members: []string

Literal :: STRING
         | INT
         | FLOAT
  * Value: interface{}

TemplateLiteral :: BACKTICK ((LCURLY Expression RCURLY) | any)* BACKTICK
  * Parts: [](Expression | string)

Context :: COMMENT? CONTEXT Selector LCURLY (UseStatement | ContextItem)* RCURLY
  * Name: string
  * Remotes: []UseStatement
  * Items: []ContextItem
  * Comment: string

ContextItemSet :: ContextItem*
  * Items: []ContextItem

ContextItem :: (ContextObject | ContextObjectMethod | ContextMethod | FunctionExpression)
  * Init: (ContextObject | ContextObjectMethod | ContextMethod | FunctionExpression)

UseStatement :: USE STRING
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

FieldStatement :: COMMENT? FieldAssignmentExpression
                | COMMENT? EnumExpression
                | COMMENT? FieldExpression
  * Init: (FieldAssignmentExpression | EnumExpression | FieldExpression)
  * Comment: string

FieldAssignmentExpression :: IDENT ASSIGN Expression
  * Name: string
  * Init: Expression

EnumExpression :: IDENT STRING
  * Name: string
  * Value: string

FieldExpression :: IDENT TypeExpression
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
            | TemplateLiteral
            | ArrayExpression
            | InstanceExpression
            | UnaryExpression
            | BinaryExpression
            | ObjectPattern
            | FunctionExpression
            | ValueExpression
            | LPAREN Expression RPAREN
  * Init: (Literal | TemplateLiteral | ArrayExpression | InstanceExpression | UnaryExpression | BinaryExpression | ObjectPattern | FunctionExpression | ValueExpression | Expression)

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
                | AssignmentStatement
                | IfStatement
                | WhileStatement
                | ForStatement
                | ContinueStatement
                | BreakStatement
                | SwitchBlock
                | GuardStatement
                | ReturnStatement
                | ThrowStatement
                | TryStatement
  * Init: (Expression | DeclarationStatement | AssignmentStatement | IfStatement | WhileStatement | ForStatement | ContinueStatement | BreakStatement | SwitchBlock | GuardStatement | ReturnStatement | ThrowStatement | TryStatement)

DeclarationStatement :: IDENT (COMMA IDENT)? DEFINE (Expression | TryStatement)
  * Target: string
  * SecondaryTarget: string?
  * Init: (Expression | TryStatement)

AssignmentStatement :: AssignmentTargetExpression token(IsAssignmentOperator) (Expression | TryStatement)
                    | AssignmentTargetExpression (COMMA IDENT)? token(IsAssignmentOperator) (Expression | TryStatement)
                    | AssignmentTargetExpression (INC | DEC)
  * Target: AssignmentTargetExpression
  * SecondaryTarget: string?
  * Operator: (token(IsAssignmentOperator) | INC | DEC)
  * Init: (Expression | TryStatement)

AssignmentTargetExpression :: IDENT AssignmentTargetExpressionMember*
  * Members: []AssignmentTargetExpressionMember

AssignmentTargetExpressionMember :: PERIOD IDENT
                                  | IndexExpression
  * Init: (string | IndexExpression)

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

ForCondition :: (DeclarationStatement | Expression) SEMICOLON Expression (SEMICOLON (Expression | AssignmentStatement))?
  * Init: DeclarationStatement?
  * Condition: Expression
  * Update: (Expression | AssignmentStatement)

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

TryStatement :: TRY Expression
  * Init: Expression
```
