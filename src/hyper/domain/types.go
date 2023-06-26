package domain

import (
	"github.com/hntrl/hyper/src/hyper/ast"
	"github.com/hntrl/hyper/src/hyper/symbols"
)

type ContextObjectInterface interface {
	FromNode(*Context, ast.ContextObject) (*ContextItem, error)
}
type ContextMethodInterface interface {
	FromNode(*Context, ast.ContextMethod) (*ContextItem, error)
}

type ObjectMethodReceiver interface {
	AddMethod(*Context, ast.ContextObjectMethod) error
}

type ContextItem struct {
	HostItem   symbols.ScopeValue
	RemoteItem symbols.ScopeValue
}
