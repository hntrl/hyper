package context

import (
	"github.com/hntrl/lang/language/nodes"
	"github.com/hntrl/lang/symbols"
)

// ObjectInterface represents an interface that can make classes from a ContextObject
type ObjectInterface interface {
	ObjectClassFromNode(*Context, nodes.ContextObject) (symbols.Class, error)
}

// ObjectMethodInterface represents any interface that can have methods defined on it
type ObjectMethodInterface interface {
	AddMethod(*Context, nodes.ContextObjectMethod) error
}

// MethodInterface represents an interface that can make classes from a ContextMethod
type MethodInterface interface {
	RemoteMethodClassFromNode(*Context, nodes.RemoteContextMethod) (symbols.Class, error)
	MethodClassFromNode(*Context, nodes.ContextMethod) (symbols.Class, error)
}

// ValueInterface represents an interface that can make objects from a ContextObject
type ValueInterface interface {
	ValueFromNode(*Context, nodes.ContextObject) (symbols.ValueObject, error)
}

type RemoteObjectInterface interface {
	Export() (symbols.Object, error)
}
