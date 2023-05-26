package interfaces

import (
	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/stdlib"
	"github.com/hntrl/hyper/internal/symbols"
)

type FileType struct {
	Name     string
	Private  bool
	Comment  string
	Lazyload bool
	Allowed  []stdlib.MimeType
	Sizes    []stdlib.Dimension
}

func (f FileType) ClassName() string {
	return f.Name
}
func (f FileType) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddConstructor(symbols.String{}, func(obj symbols.ValueObject) (symbols.ValueObject, error) {
		return File{parentType: f, fileId: string(obj.(symbols.StringLiteral))}, nil
	})
	return csMap
}
func (f FileType) Get(key string) (symbols.Object, error) {
	return nil, nil
}

func (f FileType) ObjectClassFromNode(ctx *context.Context, node ast.ContextObject) (symbols.Class, error) {
	table := ctx.Symbols()

	f.Name = node.Name
	f.Private = node.Private
	f.Comment = node.Comment

	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.AssignmentStatement:
			obj, err := table.ResolveValueObject(field.Init)
			if err != nil {
				return nil, err
			}
			switch field.Name {
			case "allowed":
				iterable, ok := obj.(symbols.Iterable)
				if !ok {
					return nil, symbols.NodeError(field.Init, "%T not allowed for sizes in file", obj.Class().ClassName())
				}
				if iterable.ParentType != (stdlib.MimeType{}) {
					return nil, symbols.NodeError(field.Init, "expected iterable of type mime.MimeType, got %s", iterable.ParentType.ClassName())
				}
				allowed := make([]stdlib.MimeType, len(iterable.Items))
				for idx, valueObj := range iterable.Items {
					allowed[idx] = valueObj.(stdlib.MimeType)
				}
				f.Allowed = allowed
			case "sizes":
				iterable, ok := obj.(symbols.Iterable)
				if !ok {
					return nil, symbols.NodeError(field.Init, "%T not allowed for sizes in file", obj.Class().ClassName())
				}
				if iterable.ParentType != (stdlib.Dimension{}) {
					return nil, symbols.NodeError(field.Init, "expected iterable of type units.Dimension, got %s", iterable.ParentType.ClassName())
				}
				sizes := make([]stdlib.Dimension, len(iterable.Items))
				for idx, valueObj := range iterable.Items {
					sizes[idx] = valueObj.(stdlib.Dimension)
				}
				f.Sizes = sizes
			case "lazyload":
				if boolLit, ok := obj.(symbols.BooleanLiteral); ok {
					f.Lazyload = bool(boolLit)
				} else {
					return nil, symbols.NodeError(field.Init, "%T not allowed for lazyload in file", obj.Class().ClassName())
				}
			default:
				return nil, symbols.NodeError(field, "unrecognized assignment %s in file", field.Name)
			}
		default:
			return nil, symbols.NodeError(field, "%T not allowed in file", item)
		}
	}
	return f, nil
}

func (f FileType) Export() (symbols.Object, error) {
	return f, nil
}

type File struct {
	parentType FileType
	fileId     string `hash:"ignore"`
	// body       io.ReadCloser `hash:"ignore"`
}

func (f File) Class() symbols.Class {
	return f.parentType
}
func (f File) Value() interface{} {
	return f.fileId
}
func (f File) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, f)
}
func (f File) Get(key string) (symbols.Object, error) {
	methods := map[string]symbols.Object{
		"toString": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{},
			Returns:   symbols.String{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				return symbols.StringLiteral(f.fileId), nil
			},
		}),
	}
	return methods[key], nil
}
