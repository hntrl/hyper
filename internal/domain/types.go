package domain

import (
	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/symbols"
)

type ContextObjectInterface interface {
	FromNode(*Context, ast.ContextObject) (ContextItem, error)
}
type ContextMethodInterface interface {
	FromNode(*Context, ast.ContextMethod) (ContextItem, error)
}

type ObjectMethodReceiver interface {
	AddMethod(*Context, ast.ContextObjectMethod) error
}

type ContextItem struct {
	HostItem   symbols.ScopeValue
	RemoteItem symbols.ScopeValue
}
