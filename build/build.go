package build

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hntrl/lang/language"
	"github.com/hntrl/lang/language/nodes"
	"github.com/hntrl/lang/resource"
	"github.com/pkg/errors"
)

// FIXME: the paradigm of leaving objects unresolved until they are used is not
// ideal -- semantic errors should be highlighted in imports even if it's
// irrelevant to the current context

// BuildContext represents the top-level structure of context usage
type BuildContext struct {
	packages   map[string]Object
	imports    map[string]*Context
	interfaces map[string]Class
	resources  map[string]resource.Resource
}

func NewBuildContext() *BuildContext {
	return &BuildContext{
		packages: make(map[string]Object),
		imports:  make(map[string]*Context),
		interfaces: map[string]Class{
			"type": &Type{},
		},
		resources: make(map[string]resource.Resource),
	}
}

func (ctx *BuildContext) GetPackage(pkg string) (interface{}, error) {
	if stdPkg := ctx.packages[pkg]; stdPkg != nil {
		return stdPkg, nil
	} else if cachedImport := ctx.imports[pkg]; cachedImport != nil {
		return cachedImport, nil
	} else {
		innerManifestTree, err := language.ParseFromFile(pkg)
		if err != nil {
			return nil, fmt.Errorf("cannot import %s: \n%s", pkg, err.Error())
		}
		innerCtx, err := prepareContext(ctx, pkg, *innerManifestTree)
		if err != nil {
			return nil, fmt.Errorf("cannot import %s: \n%s", pkg, err.Error())
		}
		return innerCtx, nil
	}
}

func (ctx *BuildContext) RegisterPackage(key string, obj Object) {
	ctx.packages[key] = obj
}
func (ctx *BuildContext) RegisterInterface(key string, class Class) {
	ctx.interfaces[key] = class
}

// Context represents a single context as defined in a manifest
type Context struct {
	Name     string
	filePath string
	buildCtx *BuildContext
	manifest nodes.Manifest

	// Represents the selectors that are available in the context
	selectors map[string]Object
	// Represents the object cache to be evaluated as required
	unresolvedItems map[string]nodes.Node
	// Represents the objects defined in the context
	Items map[string]Object
}

func NewContext(buildCtx *BuildContext, path string, node nodes.Manifest) (*Context, error) {
	ctx, err := prepareContext(buildCtx, path, node)
	if err != nil {
		return nil, err
	}
	// Don't need to use ctx.evaluateAllItems since that context is added to the build context on creation

	for _, innerCtx := range buildCtx.imports {
		err = innerCtx.evaluateAllItems()
		if err != nil {
			if innerCtx != ctx {
				return nil, fmt.Errorf("cannot import %s: %s", innerCtx.Name, err.Error())
			} else {
				return nil, err
			}
		}
	}
	// for _, innerCtx := range buildCtx.imports {
	// 	err = innerCtx.evaluateMethods()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return ctx, nil
}

// Initializes a context without resolving all of the objects
func prepareContext(buildCtx *BuildContext, path string, node nodes.Manifest) (*Context, error) {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctx := Context{
		Name:     node.Context.Name,
		filePath: path,
		buildCtx: buildCtx,
		manifest: node,
		selectors: map[string]Object{
			"String":   String{},
			"Double":   Double{},
			"Float":    Float{},
			"Int":      Integer{},
			"Bool":     Boolean{},
			"Date":     Date{},
			"DateTime": DateTime{},
			// "deprecated_PrimaryKey": ,
			"print": NewFunction(FunctionOptions{
				Arguments: []Class{
					GenericObject{},
				},
				Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
					fmt.Println(args[0])
					return nil, nil
				},
			}),
			"len": NewFunction(FunctionOptions{
				Arguments: []Class{
					// FIXME: this means arguments skip validation, but it should be
					// Indexable. not a good way to put in arguments but it has to since
					// it's an interface
					GenericObject{},
				},
				Returns: Integer{},
				Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
					if indexable, ok := args[0].(Indexable); ok {
						return IntegerLiteral(indexable.Len()), nil
					}
					return nil, fmt.Errorf("cannot get length of %s", args[0].Class().ClassName())
				},
			}),
			"emit": NewFunction(FunctionOptions{
				Arguments: []Class{
					// FIXME: same problem as len(), but instead of indexable it should be
					// classes.EventInstance
					GenericObject{},
				},
				Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
					return nil, nil
				},
			}),
			"deprecated_GenericID": NewFunction(FunctionOptions{
				Arguments: []Class{
					Integer{},
				},
				Returns: String{},
				Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
					return StringLiteral(strconv.Itoa(seededRand.Int())[0:args[0].(IntegerLiteral)]), nil
				},
			}),
		},
		unresolvedItems: make(map[string]nodes.Node),
		Items:           make(map[string]Object),
	}
	buildCtx.imports[path] = &ctx

	for _, objectNode := range node.Context.Objects {
		switch node := objectNode.(type) {
		case nodes.ContextObject:
			ctx.unresolvedItems[node.Name] = node
		case nodes.ContextMethod:
			ctx.unresolvedItems[node.Name] = node
		}
	}
	for _, importStatement := range node.Imports {
		err := ctx.Import(importStatement.Package)
		if err != nil {
			return nil, err
		}
	}
	for _, objectNode := range node.Context.Objects {
		if fnExpr, ok := objectNode.(nodes.FunctionExpression); ok {
			symbols := ctx.Symbols()
			fn, err := symbols.ResolveFunctionBlock(fnExpr.Body, &GenericObject{})
			if err != nil {
				return nil, err
			}
			ctx.selectors[fnExpr.Name] = fn
		}
	}
	return &ctx, nil
}

func (ctx *Context) evaluateItem(key string) error {
	objectNode := ctx.unresolvedItems[key]
	switch node := objectNode.(type) {
	case nodes.ContextObject:
		obj, err := ctx.resolveObject(node)
		if err != nil {
			return err
		}
		ctx.Items[key] = obj
	case nodes.ContextMethod:
		method, err := ctx.resolveMethod(node)
		if err != nil {
			return err
		}
		ctx.Items[key] = method
	}
	delete(ctx.unresolvedItems, key)
	return nil
}
func (ctx *Context) evaluateAllItems() error {
	for key, node := range ctx.unresolvedItems {
		if _, ok := node.(nodes.ContextObject); ok {
			err := ctx.evaluateItem(key)
			if err != nil {
				return err
			}
		}
	}
	for _, objectNode := range ctx.manifest.Context.Objects {
		switch node := objectNode.(type) {
		case nodes.ContextObjectMethod:
			target := ctx.Items[node.Target]
			if target == nil {
				return NodeError(node, "method target %s does not exist", node.Target)
			}
			if class, ok := target.(ObjectMethodInterface); ok {
				err := class.AddMethod(ctx, node)
				if err != nil {
					return err
				}
			} else {
				return NodeError(node, "cannot use %s as method target", node.Target)
			}
		}
	}
	for key, node := range ctx.unresolvedItems {
		if _, ok := node.(nodes.ContextMethod); ok {
			err := ctx.evaluateItem(key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ctx *Context) resolveObject(node nodes.ContextObject) (Object, error) {
	interfaceType := ctx.buildCtx.interfaces[node.Interface]
	if interfaceType == nil {
		return nil, UnknownInterfaceError(node, node.Interface)
	}

	switch targetInterface := interfaceType.(type) {
	case ObjectInterface:
		obj, err := targetInterface.ObjectClassFromNode(ctx, node)
		if err != nil {
			return nil, err
		}
		return obj, nil
	case ValueInterface:
		val, err := targetInterface.ValueFromNode(ctx, node)
		if err != nil {
			return nil, err
		}
		return val, nil
	default:
		return nil, NodeError(node, "%s cannot be created from object definition", node.Interface)
	}
}

func (ctx *Context) resolveMethod(node nodes.ContextMethod) (Class, error) {
	targetInterface := ctx.buildCtx.interfaces[node.Interface]
	if targetInterface == nil {
		return nil, UnknownInterfaceError(node, node.Interface)
	}
	if class, ok := targetInterface.(MethodInterface); ok {
		method, err := class.MethodClassFromNode(ctx, node)
		if err != nil {
			return nil, err
		}
		return method, nil
	} else {
		return nil, NodeError(node, "%s cannot be created from method definition", node.Interface)
	}
}

func (ctx *Context) Import(pkgName string) error {
	if strings.Contains(pkgName, "/") {
		pkgName = filepath.Join(filepath.Dir(ctx.filePath), pkgName)
	}

	pkgValue, err := ctx.buildCtx.GetPackage(pkgName)
	if err != nil {
		return err
	}

	switch pkg := pkgValue.(type) {
	case *Context:
		domainParts := strings.Split(pkg.Name, ".")
		if len(domainParts) == 1 {
			ctx.selectors[domainParts[0]] = pkg
		} else {
			rootDomain, ok := ctx.selectors[domainParts[0]].(Domain)
			if !ok {
				rootDomain = Domain{}
			}
			currentDomain := rootDomain
			for idx, domainPart := range domainParts[1:] {
				if idx == len(domainParts)-2 {
					currentDomain.set(domainPart, pkg)
				} else {
					if _, ok := currentDomain.Get(domainPart).(Domain); ok {
						currentDomain.set(domainPart, Domain{})
					}
					currentDomain = currentDomain.Get(domainPart).(Domain)
				}
			}
			ctx.selectors[domainParts[0]] = rootDomain
		}
	case Object:
		ctx.selectors[pkgName] = pkg
	default:
		return fmt.Errorf("cannot import %s", pkgName)
	}
	return nil
}

func (ctx *Context) Symbols() SymbolTable {
	global := make(map[string]Object)
	for key, val := range ctx.selectors {
		global[key] = val
	}
	for key, val := range ctx.Items {
		global[key] = val
	}
	return SymbolTable{
		root:      ctx,
		immutable: global,
		local:     make(map[string]*Object),
	}
}

func (ctx *Context) EvaluateTypeExpression(expr nodes.TypeExpression) (Class, error) {
	key := expr.Selector.Members[0]
	if _, ok := ctx.unresolvedItems[key]; ok {
		err := ctx.evaluateItem(key)
		if err != nil {
			return nil, err
		}
	}
	return ctx.Symbols().ResolveTypeExpression(expr)
}

func (ctx *Context) Attach() {
	for _, obj := range ctx.Items {
		if runtimeObj, ok := obj.(RuntimeNode); ok {
			err := runtimeObj.Attach(*ctx)
			if err != nil {
				ctx.Detach()
				panic(err)
			}
		}
	}
}
func (ctx *Context) Detach() {
	for _, res := range ctx.buildCtx.resources {
		res.Detach()
	}
}
func (ctx *Context) Resource(key string, vPtr interface{}) error {
	buildCtx := ctx.buildCtx
	ptr := reflect.ValueOf(vPtr)
	if ptr.Kind() != reflect.Ptr {
		return fmt.Errorf("%s is not a pointer", key)
	}
	if buildCtx.resources[key] == nil {
		if blankRes, ok := reflect.New(ptr.Type().Elem()).Interface().(resource.Resource); ok {
			res, err := blankRes.Attach()
			if err != nil {
				return err
			}
			buildCtx.resources[key] = res
		} else {
			return fmt.Errorf("%T is not a resource", vPtr)
		}
	}
	if ptr.Type().Elem() != reflect.TypeOf(buildCtx.resources[key]) {
		return fmt.Errorf("resource %s is not of type %s", key, ptr.Type().Elem())
	}
	ptr.Elem().Set(reflect.ValueOf(buildCtx.resources[key]))
	return nil
}

func (ctx *Context) Get(key string) Object {
	if _, ok := ctx.unresolvedItems[key]; ok {
		err := ctx.evaluateItem(key)
		if err != nil {
			// FIXME: this is a hack for imported objects without creating import
			// cycles. Having properties accessed this way is probably fine, but it
			// shouldn't panic if there's an error
			panic(errors.Errorf("cannot resolve %s in %s: %s", key, ctx.Name, err.Error()))
		}
	}
	return ctx.Items[key]
}

type Domain map[string]Object

func (d Domain) Get(key string) Object {
	return d[key]
}
func (d Domain) set(key string, obj Object) error {
	d[key] = obj
	return nil
}
