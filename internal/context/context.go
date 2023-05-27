package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/stdlib"
	"github.com/hntrl/hyper/internal/symbols"
)

type ContextBuilder struct {
	RootContext string
	Contexts    map[string]*Context
	Interfaces  map[string]interface{}
	Selectors   map[string]symbols.Object
}

func NewContextBuilder() *ContextBuilder {
	bd := &ContextBuilder{
		Contexts:   make(map[string]*Context),
		Interfaces: make(map[string]interface{}),
		Selectors:  make(map[string]symbols.Object),
	}
	bd.registerDefaults()
	return bd
}

func (bd *ContextBuilder) ParseContext(node ast.Manifest, path string) (*Context, error) {
	bd.RootContext = path
	targetCtx, err := bd.addContext(node, path)
	if err != nil {
		return nil, err
	}
	for _, ctx := range bd.Contexts {
		for key := range ctx.unresolvedItems {
			err := ctx.resolveItem(key)
			if err != nil {
				return nil, fmt.Errorf("cannot import %s: %s", ctx.Name, err.Error())
			}
		}
		err := ctx.resolveObjectMethods()
		if err != nil {
			return nil, fmt.Errorf("cannot import %s: %s", ctx.Name, err.Error())
		}
	}
	return targetCtx, nil
}

func (bd *ContextBuilder) addContext(node ast.Manifest, path string) (*Context, error) {
	ctx := &Context{
		Name:            node.Context.Name,
		Path:            path,
		Items:           make(map[string]symbols.Object),
		Selectors:       make(map[string]symbols.Object),
		UsedPackages:    make([]string, 0),
		Manifest:        node,
		imports:         make(map[string]symbols.Object),
		unresolvedItems: make(map[string]ast.Node),
		builder:         bd,
	}
	for key, obj := range bd.Selectors {
		ctx.Selectors[key] = obj
	}
	bd.Contexts[path] = ctx

	items := node.Context.Items
	for _, useStatement := range node.Context.Remotes {
		remotePath := filepath.Join(filepath.Dir(path), useStatement.Source)
		itemSet, err := ParseContextItemSetFromFile(remotePath)
		if err != nil {
			return nil, err
		}
		items = append(items, itemSet.Items...)
	}
	ctx.Manifest.Context.Items = items

	for _, nodeItem := range items {
		switch node := nodeItem.Init.(type) {
		case ast.ContextObject:
			ctx.unresolvedItems[node.Name] = node
		case ast.ContextMethod:
			ctx.unresolvedItems[node.Name] = node
		case ast.RemoteContextMethod:
			ctx.unresolvedItems[node.Name] = node
		case ast.FunctionExpression:
			ctx.unresolvedItems[node.Name] = node
		}
	}
	for _, importStatement := range node.Imports {
		obj, err := bd.GetPackage(importStatement.Package)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(filepath.Dir(path), importStatement.Package)
		ctx.UsedPackages = append(ctx.UsedPackages, path)
		switch pkg := obj.(type) {
		case *Context:
			err := addContextToImports(ctx, pkg)
			if err != nil {
				return nil, err
			}
		default:
			ctx.imports[importStatement.Package] = pkg
		}
	}
	return ctx, nil
}

func (bd *ContextBuilder) GetContext(name string) *Context {
	for innerName, ctx := range bd.Contexts {
		if name == innerName {
			return ctx
		}
	}
	return nil
}

func (bd *ContextBuilder) GetPackage(pkg string) (symbols.Object, error) {
	if stdPackage, ok := stdlib.Packages[pkg]; ok {
		return stdPackage, nil
	} else {
		rootCtx := bd.GetContext(bd.RootContext)
		path := filepath.Join(filepath.Dir(rootCtx.Path), pkg)
		if cachedImport, ok := bd.Contexts[path]; ok {
			return cachedImport, nil
		} else {
			manifest, err := ParseContextFromFile(path)
			if err != nil {
				return nil, fmt.Errorf("cannot import %s: \n%s", pkg, err.Error())
			}
			ctx, err := bd.addContext(*manifest, path)
			if err != nil {
				return nil, fmt.Errorf("cannot import %s: \n%s", pkg, err.Error())
			}
			return ctx, nil
		}
	}
}

func (bd *ContextBuilder) registerDefaults() {
	bd.RegisterInterface("type", TypeInterface{})
}

func (bd *ContextBuilder) RegisterInterface(key string, intf interface{}) {
	bd.Interfaces[key] = intf
}

func (bd *ContextBuilder) getInterface(target string) (interface{}, error) {
	intf := bd.Interfaces[target]
	if intf == nil {
		return nil, fmt.Errorf("unknown interface %s", target) //symbols.UnknownInterfaceError("node", target)
	}
	return intf, nil
}

// lmao wtf is this
func addContextToImports(ctx *Context, add *Context) error {
	remoteCtx := RemoteContext{add}
	domainParts := strings.Split(add.Name, ".")
	if len(domainParts) == 1 {
		ctx.imports[domainParts[0]] = remoteCtx
	} else {
		rootDomain, ok := ctx.imports[domainParts[0]].(Domain)
		if !ok {
			rootDomain = Domain{}
		}
		currentDomain := rootDomain
		for idx, domainPart := range domainParts[1:] {
			if idx == len(domainParts)-2 {
				currentDomain[domainPart] = remoteCtx
			} else {
				nextDomain := currentDomain[domainPart]
				if nextDomain == nil {
					nextDomain = Domain{}
				}
				currentDomain[domainPart] = nextDomain
			}
		}
		ctx.imports[domainParts[0]] = rootDomain
	}
	return nil
}

type Context struct {
	Name            string
	Path            string
	Items           map[string]symbols.Object
	Selectors       map[string]symbols.Object
	UsedPackages    []string
	Manifest        ast.Manifest
	imports         map[string]symbols.Object
	unresolvedItems map[string]ast.Node
	builder         *ContextBuilder
}

func (ctx *Context) addMethod(node ast.ContextMethod) (symbols.Object, error) {
	intfObject, err := ctx.builder.getInterface(node.Interface)
	if err != nil {
		return nil, err
	}
	methodIntf, ok := intfObject.(MethodInterface)
	if !ok {
		return nil, fmt.Errorf("%s cannot be initialized as method", node.Interface)
	}
	method, err := methodIntf.MethodClassFromNode(ctx, node)
	if err != nil {
		return nil, err
	}
	return method, nil
}
func (ctx *Context) addRemoteMethod(node ast.RemoteContextMethod) (symbols.Object, error) {
	intfObject, err := ctx.builder.getInterface(node.Interface)
	if err != nil {
		return nil, err
	}
	methodIntf, ok := intfObject.(MethodInterface)
	if !ok {
		return nil, fmt.Errorf("%s cannot be initialized as method", node.Interface)
	}
	method, err := methodIntf.RemoteMethodClassFromNode(ctx, node)
	if err != nil {
		return nil, err
	}
	return method, nil
}
func (ctx *Context) addObject(node ast.ContextObject) (symbols.Object, error) {
	intfObject, err := ctx.builder.getInterface(node.Interface)
	if err != nil {
		return nil, err
	}
	if objIntf, ok := intfObject.(ObjectInterface); ok {
		obj, err := objIntf.ObjectClassFromNode(ctx, node)
		if err != nil {
			return nil, err
		}
		return obj, nil
	} else if valIntf, ok := intfObject.(ValueInterface); ok {
		val, err := valIntf.ValueFromNode(ctx, node)
		if err != nil {
			return nil, err
		}
		return val, nil
	} else {
		return nil, fmt.Errorf("%s cannot be initialized as object", node.Interface)
	}
}
func (ctx *Context) addFunction(node ast.FunctionExpression) (symbols.Object, error) {
	symbolTable := ctx.Symbols()
	fn, err := symbolTable.ResolveFunctionBlock(node.Body, &symbols.MapObject{})
	if err != nil {
		return nil, err
	}
	return fn, err
}
func (ctx *Context) resolveItem(key string) error {
	switch node := ctx.unresolvedItems[key].(type) {
	case ast.ContextMethod:
		method, err := ctx.addMethod(node)
		if err != nil {
			return err
		}
		ctx.Items[key] = method
	case ast.RemoteContextMethod:
		method, err := ctx.addRemoteMethod(node)
		if err != nil {
			return err
		}
		ctx.Items[key] = method
	case ast.ContextObject:
		obj, err := ctx.addObject(node)
		if err != nil {
			return err
		}
		ctx.Items[key] = obj
	case ast.FunctionExpression:
		fn, err := ctx.addFunction(node)
		if err != nil {
			return err
		}
		ctx.Selectors[key] = fn
	}
	delete(ctx.unresolvedItems, key)
	return nil
}

func (ctx *Context) resolveObjectMethods() error {
	for _, node := range ctx.Manifest.Context.Items {
		if objMethodNode, ok := node.Init.(ast.ContextObjectMethod); ok {
			obj, err := ctx.Get(objMethodNode.Target)
			if err != nil {
				return err
			}
			if obj == nil {
				return symbols.NodeError(objMethodNode, "unknown target %s", objMethodNode.Target)
			}
			class, ok := obj.(ObjectMethodInterface)
			if !ok {
				return symbols.NodeError(objMethodNode, "cannot use %s as method target", objMethodNode.Target)
			}
			err = class.AddMethod(ctx, objMethodNode)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ctx *Context) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(filepath.Dir(ctx.Path), path))
}

func (ctx *Context) Symbols() symbols.SymbolTable {
	return symbols.NewSymbolTable(ctx)
}

func (ctx *Context) Get(key string) (symbols.Object, error) {
	if domain, ok := ctx.imports[key]; ok {
		return domain, nil
	}
	if _, ok := ctx.unresolvedItems[key]; ok {
		err := ctx.resolveItem(key)
		if err != nil {
			return nil, err
		}
	}
	if obj, ok := ctx.Selectors[key]; ok {
		return obj, nil
	}
	return ctx.Items[key], nil
}

// Represents a context that is being accessed from another context
type RemoteContext struct {
	Context *Context
}

func (rc RemoteContext) Get(key string) (symbols.Object, error) {
	obj, err := rc.Context.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot import %s: %s", rc.Context.Name, err.Error())
	}
	if objIntf, ok := obj.(RemoteObjectInterface); ok {
		return objIntf.Export()
	}
	return nil, nil
}

type Domain map[string]symbols.Object

func (d Domain) Get(key string) (symbols.Object, error) {
	return d[key], nil
}
