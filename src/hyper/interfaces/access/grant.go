package access

import (
	"github.com/hntrl/hyper/src/hyper/ast"
	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/hyper/symbols"
	"github.com/hntrl/hyper/src/hyper/symbols/errors"
)

type GrantInterface struct{}

func (GrantInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	if node.Private {
		return nil, errors.NodeError(node, 0, "grant cannot be private: grants aren't exported")
	}
	grant := GrantValue{
		Name:        node.Name,
		Description: node.Comment,
	}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.FieldAssignmentExpression:
			switch field.Name {
			case "name":
				nameValue, err := table.ResolveExpression(field.Init)
				if err != nil {
					return nil, err
				}
				strValue, ok := nameValue.(symbols.StringValue)
				if !ok {
					return nil, errors.NodeError(field.Init, 0, "expected String for name, got %s", nameValue.Class().Descriptors().Name)
				}
				grant.Name = string(strValue)
			case "description":
				descriptionValue, err := table.ResolveExpression(field.Init)
				if err != nil {
					return nil, err
				}
				strValue, ok := descriptionValue.(symbols.StringValue)
				if !ok {
					return nil, errors.NodeError(field.Init, 0, "expected String for description, got %s", descriptionValue.Class().Descriptors().Name)
				}
				grant.Description = string(strValue)
			default:
				return nil, errors.NodeError(field, 0, "unrecognized assignment %s in grant", field.Name)
			}
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in parameter", item)
		}
	}
	return &domain.ContextItem{
		HostItem:   grant,
		RemoteItem: nil,
	}, nil
}

var (
	Grant            = GrantClass{}
	GrantDescriptors = &symbols.ClassDescriptors{
		Name: "Grant",
	}
)

type GrantClass struct{}

func (GrantClass) Descriptors() *symbols.ClassDescriptors {
	return GrantDescriptors
}

type GrantValue struct {
	Name        string
	Description string
}

func (GrantValue) Class() symbols.Class {
	return Grant
}
func (GrantValue) Value() interface{} {
	return nil
}
