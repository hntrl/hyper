package symbols

// Object represents anything that can be referenced in scope that doesn't hold
// any state
type Object interface {
	Get(string) (Object, error)
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
	Object
	Arguments() []Class
	Returns() Class
	Call([]ValueObject, ValueObject) (ValueObject, error)
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
