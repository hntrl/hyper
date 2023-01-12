package build

import (
	"github.com/hntrl/lang/language/nodes"
)

// Object represents anything that can be referenced in scope that doesn't hold
// any state
type Object interface {
	Get(string) Object
}

// Class represents the abstract type of a ValueObject
type Class interface {
	Object
	ClassName() string
	Constructors() ConstructorMap
}

// ObjectClass represents any class that has properties to describe it
type ObjectClass interface {
	Class
	Fields() map[string]Class
}

// ComparableClass represents anything that can be compared to another object
// i.e. a == b
type ComparableClass interface {
	Class
	ComparableRules() ComparatorRules
}

// OperableClass represents anything that can be modified by a binary expression
// i.e. a + b
type OperableClass interface {
	Class
	OperatorRules() OperatorRules
}

// Method represents anything that can be called with arguments
type Method interface {
	Arguments() []Class
	Returns() Class
	Call([]ValueObject, ValueObject) (ValueObject, error)
}

// ObjectInterface represents an interface that can make classes from a ContextObject
type ObjectInterface interface {
	ObjectClassFromNode(*Context, nodes.ContextObject) (Class, error)
}

// ObjectMethodInterface represents any interface that can have methods defined on it
type ObjectMethodInterface interface {
	AddMethod(*Context, nodes.ContextObjectMethod) error
}

// MethodInterface represents an interface that can make classes from a ContextMethod
type MethodInterface interface {
	MethodClassFromNode(*Context, nodes.ContextMethod) (Class, error)
}

// ValueInterface represents an interface that can make objects from a ContextObject
type ValueInterface interface {
	ValueFromNode(*Context, nodes.ContextObject) (ValueObject, error)
}

// ValueObject represents the stateful representation of a Class
type ValueObject interface {
	Object
	Class() Class
	Value() interface{}
	Set(string, ValueObject) error
}

// Indexable represents anything that can be accessed using indices
// i.e. a[0] or a[0:2]
type Indexable interface {
	ValueObject
	GetIndex(int) (ValueObject, error)
	SetIndex(int, ValueObject) error
	Range(int, int) (Indexable, error)
	Len() int
}

// Exportable represents the object a class assumes when being imported from an
// outside context
type Exportable interface {
	Class
	IsExported() bool
	ExportedObject() Object
}

// RuntimeNode represents anything that should be attached to the runtime process
type RuntimeNode interface {
	Attach(ctx Context) error
	Detach() error
}
