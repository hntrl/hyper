package interfaces

import (
	"fmt"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/symbols/errors"
	"github.com/hntrl/hyper/internal/tokens"
)

type EnumInterface struct{}

func (EnumInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	enum := &Enum{
		Name:    node.Name,
		Private: node.Private,
		Comment: node.Comment,
		items:   make(map[string]EnumItem),
	}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.EnumExpression:
			enum.items[field.Name] = EnumItem{
				parentType:  enum,
				stringValue: string(field.Name),
			}
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in enum", item)
		}
	}
	if !node.Private {
		return &domain.ContextItem{
			HostItem:   enum,
			RemoteItem: enum,
		}, nil
	} else {
		return &domain.ContextItem{
			HostItem:   enum,
			RemoteItem: nil,
		}, nil
	}
}

type Enum struct {
	Name    string
	Private bool
	Comment string
	items   map[string]EnumItem `hash:"ignore"`
}

func (en Enum) Descriptors() *symbols.ClassDescriptors {
	return &symbols.ClassDescriptors{
		Name: en.Name,
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(en, func(val *EnumItem) (EnumItem, error) {
				return *val, nil
			}),
			symbols.Constructor(symbols.String, func(val symbols.StringValue) (*EnumItem, error) {
				item, ok := en.items[string(val)]
				if !ok {
					return nil, fmt.Errorf("%s not valid for %s", val.Value(), en.Name)
				}
				return &item, nil
			}),
		},
		Comparators: symbols.ClassComparatorSet{
			symbols.Comparator(en, tokens.EQUALS, func(a, b EnumItem) (bool, error) {
				return a.Value() == b.Value(), nil
			}),
			symbols.Comparator(en, tokens.NOT_EQUALS, func(a, b EnumItem) (bool, error) {
				return a.Value() != b.Value(), nil
			}),
		},
	}
}

type EnumItem struct {
	parentType  *Enum
	stringValue string `hash:"ignore"`
}

func (ei EnumItem) Class() symbols.Class {
	return ei.parentType
}
func (ei EnumItem) Value() interface{} {
	return ei.stringValue
}
