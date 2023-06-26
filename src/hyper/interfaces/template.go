package interfaces

import (
	"bytes"
	"path"

	"github.com/hntrl/hyper/src/hyper/ast"
	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/hyper/symbols"
	"github.com/hntrl/hyper/src/hyper/symbols/errors"
	"github.com/kataras/blocks"
)

type TemplateInterface struct{}

func (TemplateInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	if node.Private {
		return nil, errors.NodeError(node, 0, "template cannot be private: templates aren't exported")
	}

	builder := TemplateBuilder{
		Name:    node.Name,
		Comment: node.Comment,
	}
	blocks := blocks.New(path.Join(path.Dir(string(ctx.Path)), "./templates"))
	err := blocks.Load()
	if err != nil {
		return nil, err
	}
	builder.blocks = blocks

	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.FieldAssignmentExpression:
			switch field.Name {
			case "src":
				fileNameValue, err := table.ResolveExpression(field.Init)
				if err != nil {
					return nil, err
				}
				strValue, ok := fileNameValue.(symbols.StringValue)
				if !ok {
					return nil, errors.NodeError(field.Init, 0, "expected String for src, got %s", fileNameValue.Class().Descriptors().Name)
				}
				builder.fileName = string(strValue)
			case "layout":
				layoutNameValue, err := table.ResolveExpression(field.Init)
				if err != nil {
					return nil, err
				}
				strValue, ok := layoutNameValue.(symbols.StringValue)
				if !ok {
					return nil, errors.NodeError(field.Init, 0, "expected String for layout, got %s", layoutNameValue.Class().Descriptors().Name)
				}
				builder.layoutName = string(strValue)
			default:
				return nil, errors.NodeError(field, 0, "unrecognized assignment %s in template", field.Name)
			}
		case ast.FieldExpression:
			switch field.Name {
			case "type":
				typeClass, err := table.EvaluateTypeExpression(field.Init)
				if err != nil {
					return nil, err
				}
				builder.parentType = typeClass
			default:
				return nil, errors.NodeError(field, 0, "unrecognized field %s in template", field.Name)
			}
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in template", item)
		}
	}
	if builder.parentType == nil {
		return nil, errors.NodeError(node, 0, "template must have a type")
	}
	if builder.fileName == "" {
		return nil, errors.NodeError(node, 0, "template must have a `src` attribute")
	}
	if builder.layoutName == "" {
		return nil, errors.NodeError(node, 0, "template must have a `layout` attribute")
	}
	return &domain.ContextItem{
		HostItem:   builder,
		RemoteItem: nil,
	}, nil
}

type TemplateBuilder struct {
	Name       string
	Comment    string
	parentType symbols.Class
	fileName   string
	layoutName string
	blocks     *blocks.Blocks
}

func (tb TemplateBuilder) Arguments() []symbols.Class {
	return []symbols.Class{tb.parentType}
}
func (tb TemplateBuilder) Returns() symbols.Class {
	return symbols.String
}
func (tb TemplateBuilder) Call(args ...symbols.ValueObject) (symbols.ValueObject, error) {
	var buffer bytes.Buffer
	err := tb.blocks.ExecuteTemplate(&buffer, tb.fileName, tb.layoutName, args[0].Value())
	if err != nil {
		return nil, err
	}
	return symbols.StringValue(buffer.String()), nil
}
