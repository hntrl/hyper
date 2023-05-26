package access

import (
	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/symbols"
)

type Grant struct{}

func (gr Grant) ClassName() string {
	return "Grant"
}
func (gr Grant) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}
func (gr Grant) Get(key string) (symbols.Object, error) {
	return nil, nil
}

func (gr Grant) ValueFromNode(ctx *context.Context, node ast.ContextObject) (symbols.ValueObject, error) {
	table := ctx.Symbols()
	lit := GrantLiteral{}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.AssignmentStatement:
			expr, err := table.ResolveValueObject(field.Init)
			if err != nil {
				return nil, err
			}
			if strLit, ok := expr.(symbols.StringLiteral); ok {
				switch field.Name {
				case "name":
					lit.Name = string(strLit)
				case "description":
					lit.Description = string(strLit)
				default:
					return nil, symbols.NodeError(field, "unknown field %s in grant", field.Name)
				}
			} else {
				return nil, symbols.NodeError(field, "parsing: %s not allowed in grant", expr.Class().ClassName())
			}
		default:
			return nil, symbols.NodeError(field, "parsing: %T not allowed in grant", field)
		}
	}
	return lit, nil
}

type GrantLiteral struct {
	Name        string
	Description string
}

func (gr GrantLiteral) Class() symbols.Class {
	return Grant{}
}
func (gr GrantLiteral) Value() interface{} {
	return nil
}
func (gr GrantLiteral) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, gr)
}
func (gr GrantLiteral) Get(key string) (symbols.Object, error) {
	return nil, nil
}
