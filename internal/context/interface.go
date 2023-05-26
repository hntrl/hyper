package context

import (
	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/symbols"
)

// ObjectInterface represents an interface that can make classes from a ContextObject
type ObjectInterface interface {
	ObjectClassFromNode(*Context, ast.ContextObject) (symbols.Class, error)
}

// ObjectMethodInterface represents any interface that can have methods defined on it
type ObjectMethodInterface interface {
	AddMethod(*Context, ast.ContextObjectMethod) error
}

// MethodInterface represents an interface that can make classes from a ContextMethod
type MethodInterface interface {
	RemoteMethodClassFromNode(*Context, ast.RemoteContextMethod) (symbols.Class, error)
	MethodClassFromNode(*Context, ast.ContextMethod) (symbols.Class, error)
}

// ValueInterface represents an interface that can make objects from a ContextObject
type ValueInterface interface {
	ValueFromNode(*Context, ast.ContextObject) (symbols.ValueObject, error)
}

type RemoteObjectInterface interface {
	Export() (symbols.Object, error)
}
