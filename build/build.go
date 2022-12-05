package build

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/hntrl/lang/language"
	"github.com/hntrl/lang/language/nodes"
	"github.com/hntrl/lang/resource"
)

// FIXME: the paradigm of leaving objects unresolved until they are used is not
// ideal -- semantic errors should be highlighted in imports even if it's
// irrelevant to the current context

// BuildContext represents the top-level structure of a build process
type BuildContext struct {
	packages  map[string]Object
	imports   map[string]*Context
	classes   map[string]Class
	resources map[string]resource.Resource
}

func NewBuildContext() *BuildContext {
	return &BuildContext{
		packages: make(map[string]Object),
		imports:  make(map[string]*Context),
		classes: map[string]Class{
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
		innerCtx, err := NewContext(ctx, pkg, *innerManifestTree)
		if err != nil {
			return nil, fmt.Errorf("cannot import %s: \n%s", pkg, err.Error())
		}
		return innerCtx, nil
	}
}

func (ctx *BuildContext) RegisterPackage(key string, obj Object) {
	ctx.packages[key] = obj
}
func (ctx *BuildContext) RegisterClass(key string, class Class) {
	ctx.classes[key] = class
}

// Context represents a single context as defined in a manifest
type Context struct {
	Name     string
	filePath string
	buildCtx *BuildContext

	// Represents the selectors that are available in the context
	selectors map[string]Object
	// Represents the object cache to be evaluated as required
	unresolvedObjects map[string]nodes.ContextObject
	// Represents the objects defined in the context
	objects map[string]Object
}

func NewContext(buildCtx *BuildContext, path string, node nodes.Manifest) (*Context, error) {
	ctx := Context{
		Name:     node.Context.Name,
		filePath: path,
		buildCtx: buildCtx,
		selectors: map[string]Object{
			"String":   String{},
			"Double":   Double{},
			"Float":    Float{},
			"Int":      Integer{},
			"Bool":     Boolean{},
			"Date":     Date{},
			"DateTime": DateTime{},
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
					if indexable, ok := args[0].(Indexable[ValueObject]); ok {
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
		},
		unresolvedObjects: make(map[string]nodes.ContextObject),
		objects:           make(map[string]Object),
	}
	for _, obj := range node.Context.Objects {
		if node, ok := obj.(nodes.ContextObject); ok {
			ctx.unresolvedObjects[node.Name] = node
		}
	}
	buildCtx.imports[path] = &ctx
	err := ctx.parseManifestTree(node)
	if err != nil {
		return nil, err
	}
	return &ctx, err
}

func (ctx *Context) evaluateObject(key string) error {
	node := ctx.unresolvedObjects[key]
	classType := ctx.buildCtx.classes[node.Class]
	if classType != nil {
		if objectClass, ok := classType.(ObjectInterface); ok {
			obj, err := objectClass.ObjectClassFromNode(ctx, node)
			if err != nil {
				return err
			}
			ctx.objects[key] = obj
		} else if valueClass, ok := classType.(ValueInterface); ok {
			val, err := valueClass.ValueFromNode(ctx, node)
			if err != nil {
				return err
			}
			ctx.objects[key] = val
		} else {
			return NodeError(node, "%s cannot be created from object definition", node.Class)
		}
		delete(ctx.unresolvedObjects, key)
		return nil
	} else {
		return UnknownInterfaceError(node, node.Class)
	}
}

func (ctx *Context) parseManifestTree(node nodes.Manifest) error {
	for _, importStatement := range node.Imports {
		err := ctx.Import(importStatement.Package)
		if err != nil {
			return err
		}
	}
	for key := range ctx.unresolvedObjects {
		err := ctx.evaluateObject(key)
		if err != nil {
			return err
		}
	}
	for _, objectNode := range node.Context.Objects {
		if methodNode, ok := objectNode.(nodes.ContextMethod); ok {
			object := ctx.buildCtx.classes[methodNode.Class]
			if object != nil {
				if class, ok := object.(MethodInterface); ok {
					method, err := class.MethodClassFromNode(ctx, methodNode)
					if err != nil {
						return err
					}
					ctx.objects[methodNode.Name] = method
				} else {
					return NodeError(methodNode, "%s cannot be created from method definition", methodNode.Class)
				}
			} else {
				return UnknownInterfaceError(methodNode, methodNode.Class)
			}
		} else if objectMethodNode, ok := objectNode.(nodes.ContextObjectMethod); ok {
			object := ctx.objects[objectMethodNode.Target]
			if object != nil {
				if class, ok := object.(ObjectMethodInterface); ok {
					err := class.AddMethod(ctx, objectMethodNode)
					if err != nil {
						return err
					}
				} else {
					return NodeError(objectMethodNode, "cannot use %s as method target", objectMethodNode.Target)
				}
			} else {
				return NodeError(objectMethodNode, "method target %s does not exist", objectMethodNode.Target)
			}
		}
	}
	return nil
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
		ctx.selectors[domainParts[0]] = currentDomain
		return nil
	case Object:
		ctx.selectors[pkgName] = pkg
		return nil
	default:
		return fmt.Errorf("cannot import %s", pkgName)
	}
}

func (ctx Context) Symbols() SymbolTable {
	global := make(map[string]Object)
	for key, val := range ctx.selectors {
		global[key] = val
	}
	for key, val := range ctx.objects {
		global[key] = val
	}
	return SymbolTable{
		immutable: global,
		local:     make(map[string]Object),
	}
}

func (ctx *Context) EvaluateTypeExpression(expr nodes.TypeExpression) (Class, error) {
	key := strings.Join(expr.Selector.Members, ".")
	if _, ok := ctx.unresolvedObjects[key]; ok {
		err := ctx.evaluateObject(key)
		if err != nil {
			return nil, err
		}
	}
	return ctx.Symbols().ResolveTypeExpression(expr)
}

func (ctx *Context) Attach() {
	for _, obj := range ctx.objects {
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
	if _, ok := ctx.unresolvedObjects[key]; ok {
		err := ctx.evaluateObject(key)
		if err != nil {
			// FIXME: this is a hack for imported objects without creating import
			// cycles. Having properties accessed this way is probably fine, but it
			// shouldn't panic if there's an error
			panic(err)
		}
	}
	return ctx.objects[key]
}

type Domain map[string]Object

func (d Domain) Get(key string) Object {
	return d[key]
}
func (d Domain) set(key string, obj Object) error {
	d[key] = obj
	return nil
}
