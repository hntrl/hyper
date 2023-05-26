package interfaces

import (
	"bytes"
	"fmt"
	"path"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/kataras/blocks"
)

type Template struct {
	Name       string
	Private    bool
	Comment    string
	ParentType symbols.ObjectClass
	FileName   string
	LayoutName string
	Blocks     *blocks.Blocks
}

func (tm Template) ClassName() string {
	return tm.Name
}
func (tm Template) Fields() map[string]symbols.Class {
	return tm.ParentType.Fields()
}
func (tm Template) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}
func (tm Template) Get(key string) (symbols.Object, error) {
	return nil, nil
}

func (tm Template) Arguments() []symbols.Class {
	return []symbols.Class{tm.ParentType}
}
func (tm Template) Returns() symbols.Class {
	return symbols.String{}
}
func (tm Template) Call(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
	var buffer bytes.Buffer
	err := tm.Blocks.ExecuteTemplate(&buffer, tm.FileName, tm.LayoutName, args[0].Value())
	if err != nil {
		return nil, err
	}
	return symbols.StringLiteral(buffer.String()), nil
}

func (tm Template) ObjectClassFromNode(ctx *context.Context, node ast.ContextObject) (symbols.Class, error) {
	table := ctx.Symbols()
	templateValue := Template{
		Name:    node.Name,
		Private: node.Private,
		Comment: node.Comment,
	}

	blocks := blocks.New(path.Join(path.Dir(ctx.Path), "./templates"))
	err := blocks.Load()
	if err != nil {
		return nil, err
	}
	templateValue.Blocks = blocks

	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.AssignmentStatement:
			expr, err := table.ResolveValueObject(field.Init)
			if err != nil {
				return nil, err
			}
			strLit, ok := expr.(symbols.StringLiteral)
			if !ok {
				return nil, symbols.NodeError(field.Init, "cannot use %T as assignment in email template", expr)
			}

			if field.Name == "src" {
				templateValue.FileName = string(strLit)
			} else if field.Name == "layout" {
				templateValue.LayoutName = string(strLit)
			} else {
				return nil, symbols.NodeError(field, "unknown key %s for assignment in email template", field.Name)
			}
		case ast.TypeStatement:
			if field.Name == "type" {
				expr, err := table.ResolveTypeExpression(field.Init)
				if err != nil {
					return nil, err
				}
				objectClass, ok := expr.(symbols.ObjectClass)
				if !ok {
					return nil, symbols.NodeError(field, "email template type must be an object class")
				}
				templateValue.ParentType = objectClass
			} else {
				return nil, symbols.NodeError(field, "unknown type %s in email template", field.Name)
			}
		default:
			return nil, fmt.Errorf("parsing: %T not allowed in email template", item)
		}
	}
	if templateValue.ParentType == nil {
		return nil, symbols.NodeError(node, "missing type in email template")
	}
	if templateValue.FileName == "" {
		return nil, symbols.NodeError(node, "missing src in email template")
	}
	if templateValue.LayoutName == "" {
		return nil, symbols.NodeError(node, "missing layout in email template")
	}
	return templateValue, nil
}
