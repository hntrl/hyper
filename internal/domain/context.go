package domain

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/stdlib"
	"github.com/hntrl/hyper/internal/symbols"
)

type ContextPath string

type ContextBuilder struct {
	hostContextPath ContextPath
	contexts        map[ContextPath]*Context
	interfaces      map[string]interface{}
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		contexts:   make(map[ContextPath]*Context),
		interfaces: make(map[string]interface{}),
	}
}

func (bd *ContextBuilder) ParseContext(node ast.Manifest, path string) (*Context, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	bd.hostContextPath = ContextPath(absPath)
	err = bd.addContext(node, absPath)
	if err != nil {
		return nil, err
	}
	for _, ctx := range bd.contexts {
		for key := range ctx.unresolvedItems {
			err := ctx.resolveItem(key)
			if err != nil {
				return nil, fmt.Errorf("cannot import %s: %s", ctx.Identifier, err.Error())
			}
		}
		err := ctx.propagateObjectMethods()
		if err != nil {
			return nil, fmt.Errorf("cannot import %s: %s", ctx.Identifier, err.Error())
		}
	}
	return bd.contexts[bd.hostContextPath], nil
}
func (bd *ContextBuilder) addContext(node ast.Manifest, path string) error {
	cwd := filepath.Dir(path)
	for _, useStatement := range node.Context.Remotes {
		remotePath := filepath.Join(cwd, useStatement.Source)
		itemSet, err := ParseContextItemSetFromFile(remotePath)
		if err != nil {
			return err
		}
		node.Context.Items = append(node.Context.Items, itemSet.Items...)
	}
	ctx := &Context{
		Identifier:      node.Context.Name,
		Path:            ContextPath(path),
		Items:           make(map[string]ContextItem),
		unresolvedItems: make(map[string]ast.Node),
		selectors:       make(map[string]symbols.ScopeValue),
		manifestNode:    node,
		builder:         bd,
	}
	bd.contexts[ctx.Path] = ctx
	for _, contextItem := range node.Context.Items {
		switch node := contextItem.Init.(type) {
		case ast.ContextObject:
			ctx.unresolvedItems[node.Name] = node
		case ast.ContextMethod:
			ctx.unresolvedItems[node.Name] = node
		case ast.FunctionExpression:
			ctx.unresolvedItems[node.Name] = node
		}
	}
	for _, importStatement := range node.Imports {
		err := ctx.ImportPackage(importStatement.Source)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bd *ContextBuilder) RegisterInterface(key string, val interface{}) {
	bd.interfaces[key] = val
}
func (bd *ContextBuilder) HostContext() *Context {
	return bd.GetContextByPath(string(bd.hostContextPath))
}
func (bd *ContextBuilder) GetContextByPath(path string) *Context {
	return bd.contexts[ContextPath(path)]
}
func (bd *ContextBuilder) GetContextByIdentifier(identifier string) *Context {
	for _, ctx := range bd.contexts {
		if ctx.Identifier == identifier {
			return ctx
		}
	}
	return nil
}

type Context struct {
	Identifier      string
	Path            ContextPath
	Items           map[string]ContextItem
	unresolvedItems map[string]ast.Node
	selectors       map[string]symbols.ScopeValue
	manifestNode    ast.Manifest
	builder         *ContextBuilder
}

func (ctx *Context) ImportPackage(source string) error {
	if stdlibPackage, ok := stdlib.Packages[source]; ok {
		ctx.selectors[source] = stdlibPackage
		return nil
	}
	absPath, err := filepath.Abs(filepath.Join(filepath.Dir(string(ctx.Path)), source))
	if err != nil {
		return err
	}
	if existingContext := ctx.builder.GetContextByPath(absPath); existingContext != nil {
		addContextToSelectors(ctx, existingContext)
		return nil
	}
	manifest, err := ParseContextFromFile(absPath)
	if err != nil {
		return err
	}
	if existingContext := ctx.builder.GetContextByIdentifier(manifest.Context.Name); existingContext != nil {
		return fmt.Errorf("cannot import %s: context with identifier %s is already imported", source, manifest.Context.Name)
	}
	return ctx.builder.addContext(*manifest, absPath)
}

func (ctx *Context) addObject(node ast.ContextObject) error {
	targetInterface, ok := ctx.builder.interfaces[node.Interface]
	if !ok {
		return fmt.Errorf("cannot find interface %s", node.Interface)
	}
	contextObjectInterface, ok := targetInterface.(ContextObjectInterface)
	if !ok {
		return fmt.Errorf("interface %s does not implement ContextObjectInterface", node.Interface)
	}
	contextItem, err := contextObjectInterface.FromNode(ctx, node)
	if err != nil {
		return err
	}
	ctx.Items[node.Name] = contextItem
	return nil
}
func (ctx *Context) addMethod(node ast.ContextMethod) error {
	targetInterface, ok := ctx.builder.interfaces[node.Interface]
	if !ok {
		return fmt.Errorf("cannot find interface %s", node.Interface)
	}
	contextMethodInterface, ok := targetInterface.(ContextMethodInterface)
	if !ok {
		return fmt.Errorf("interface %s does not implement ContextMethodInterface", node.Interface)
	}
	contextItem, err := contextMethodInterface.FromNode(ctx, node)
	if err != nil {
		return err
	}
	ctx.Items[node.Name] = contextItem
	return nil
}
func (ctx *Context) addFunction(node ast.FunctionExpression) error {
	symbolTable := ctx.Symbols()
	fn, err := symbolTable.ResolveFunctionBlock(node.Body)
	if err != nil {
		return err
	}
	ctx.selectors[node.Name] = fn
	return nil
}
func (ctx *Context) resolveItem(key string) error {
	switch node := ctx.unresolvedItems[key].(type) {
	case ast.ContextMethod:
		err := ctx.addMethod(node)
		if err != nil {
			return err
		}
	case ast.ContextObject:
		err := ctx.addObject(node)
		if err != nil {
			return err
		}
	case ast.FunctionExpression:
		err := ctx.addFunction(node)
		if err != nil {
			return err
		}
	}
	delete(ctx.unresolvedItems, key)
	return nil
}

func (ctx *Context) propagateObjectMethods() error {
	for _, node := range ctx.manifestNode.Context.Items {
		if objectMethodNode, ok := node.Init.(ast.ContextObjectMethod); ok {
			target, ok := ctx.Items[objectMethodNode.Target]
			if !ok {
				return fmt.Errorf("cannot find target %s", objectMethodNode.Target)
			}
			methodReceiver, ok := target.HostItem.(ObjectMethodReceiver)
			if !ok {
				return fmt.Errorf("cannot add method %s to %s: target does not implement ObjectMethodReceiver", objectMethodNode.Name, objectMethodNode.Target)
			}
			err := methodReceiver.AddMethod(ctx, objectMethodNode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ctx *Context) Symbols() *symbols.SymbolTable {
	return symbols.NewSymbolTable(ctx)
}

func (ctx *Context) Get(key string) (symbols.ScopeValue, error) {
	if _, ok := ctx.unresolvedItems[key]; ok {
		err := ctx.resolveItem(key)
		if err != nil {
			return nil, err
		}
	}
	if selector, ok := ctx.selectors[key]; ok {
		return selector, nil
	}
	return ctx.Items[key].HostItem, nil
}

// RemoteContext is the object that gets added to a context when being imported
// from another context. Instead of accessing items that aren't relevant to the
// host context (the `HostItems`), this object will access the remote items instead
type RemoteContext struct {
	inner *Context
}

func (rc RemoteContext) Get(key string) (symbols.ScopeValue, error) {
	obj, err := rc.inner.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot import %s: %s", rc.inner.Identifier, err.Error())
	}
	if contextItem, ok := obj.(ContextItem); ok {
		return contextItem.RemoteItem, nil
	}
	return nil, nil
}

// Domain is the ambiguous object that is used to separate contexts by their
// selector parts into an object that the interpreter can understand.
type Domain map[string]symbols.ScopeValue

func NewDomainFromSelectors(selectors map[string]symbols.ScopeValue) Domain {
	out := make(Domain)
	for key, val := range selectors {
		switch selector := val.(type) {
		case Domain, *Context:
			out[key] = selector
		}
	}
	return out
}

func (d Domain) Get(key string) (symbols.ScopeValue, error) {
	return d[key], nil
}
func (d Domain) AddContextBySelector(selector string, ctx *Context) error {
	selectorParts := strings.Split(selector, ".")
	if len(selectorParts) == 1 {
		d[selector] = RemoteContext{inner: ctx}
		return nil
	}
	existingItem := d[selectorParts[0]]
	if existingItem == nil {
		nextDomain := Domain{}
		nextDomain.AddContextBySelector(strings.Join(selectorParts[1:], "."), ctx)
		d[selectorParts[0]] = nextDomain
		return nil
	}
	switch target := existingItem.(type) {
	case *Context:
		return fmt.Errorf("cannot import %s: contexts cannot be a property of another context (overwritting %s)", ctx.Identifier, target.Identifier)
	case Domain:
		return target.AddContextBySelector(strings.Join(selectorParts[1:], "."), ctx)
	}
	return nil
}

func addContextToSelectors(ctx *Context, importedCtx *Context) {
	domain := NewDomainFromSelectors(ctx.selectors)
	domain.AddContextBySelector(importedCtx.Identifier, importedCtx)
	ctx.selectors = domain
}
