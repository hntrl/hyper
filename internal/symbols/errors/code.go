package errors

type Code int

// This file defines the different error codes that can be produced when
// evaluating and resolving the syntax tree. These codes act as an identifier
// that may be used to implement special handling for certain errors.
//
// Error codes shouldn't be changed -- add new ones at the end.
//
// Error codes should be fine-grained enough that the exact nature of the error
// can be determined, but coarse enough that they aren't an implementation
// detail of the symbol table.
//
// Error code names should be as brief as possible while retaining accuracy and
// distinctiveness. In most cases names should start with an adjective
// describing the nature of the error (like "invalid", "unused") and end with a
// noun identifying the relevant language object. Naming follows the convention
// that "bad" implies a problem with syntax and "invalid" implies a problem with
// types.

const (
	// InvalidSyntaxTree occurs if an invalid syntax tree is provided to the interpreter.
	InvalidSyntaxTree Code = -1

	ExpectedCallbackSignaure
)

const (
	// The zero Code value indicates an unset error code.
	_ Code = iota

	BadUnaryOperator

	BadLoopControlStatement

	InvalidCallExpression

	InvalidClass

	InvalidClassConstruction

	InvalidInstanceableTarget

	InvalidUnaryOperand

	InvalidValueExpression

	InvalidIndex

	InvalidIndexTarget

	InvalidAssignmentTarget

	InvalidIfCondition

	InvalidWhileCondition

	InvalidForCondition

	InvalidSwitchTarget

	InvalidThrowValue

	InvalidReturnType

	InvalidSpreadTarget

	InvalidArgumentLength

	InvalidOperator

	InvalidDestructuredArgument

	CannotAccessProperty

	CannotSetProperty

	CannotEnumerate

	CannotReassignImmutableValue

	CannotRedeclareValue

	CannotConstruct

	DuplicateDefaultSwitchStatements

	MissingReturn

	UnknownSelector

	UnknownProperty

	MissingProperty

	UndefinedOperator

	// Error codes that are yielded exclusively when resolving

	IndexOutOfRange

	InvalidRangeIndices

	CannotOperateNilValue

	CannotEnumerateNilValue
)
