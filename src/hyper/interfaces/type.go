package interfaces

import (
	"github.com/hntrl/hyper/src/hyper/ast"
	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/hyper/symbols"
	"github.com/hntrl/hyper/src/hyper/symbols/errors"
)

type TypeInterface struct{}

func (TypeInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	t := TypeClass{
		Name:       node.Name,
		Private:    node.Private,
		Comment:    node.Comment,
		Properties: make(map[string]symbols.Class),
	}
	if node.Extends != nil {
		extendedType, err := table.ResolveSelector(*node.Extends)
		if err != nil {
			return nil, err
		}
		extendedTypeClass, ok := extendedType.(symbols.Class)
		if !ok {
			return nil, errors.NodeError(node.Extends, 0, "cannot extend %T", extendedType)
		}
		properties := extendedTypeClass.Descriptors().Properties
		if properties == nil {
			return nil, errors.NodeError(node.Extends, 0, "cannot extend %s", extendedTypeClass.Descriptors().Name)
		}
		for k, v := range properties {
			t.Properties[k] = v.PropertyClass
		}
	}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.FieldExpression:
			class, err := table.EvaluateTypeExpression(field.Init)
			if err != nil {
				return nil, err
			}
			t.Properties[field.Name] = class
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in type", item)
		}
	}
	if !node.Private {
		return &domain.ContextItem{
			HostItem:   t,
			RemoteItem: t,
		}, nil
	} else {
		return &domain.ContextItem{
			HostItem:   t,
			RemoteItem: nil,
		}, nil
	}
}

type TypeClass struct {
	Name       string
	Private    bool
	Comment    string
	Properties map[string]symbols.Class
}

func (tc TypeClass) Descriptors() *symbols.ClassDescriptors {
	propertyMap := make(symbols.ClassPropertyMap)
	for name, class := range tc.Properties {
		propertyMap[name] = symbols.PropertyAttributes(symbols.PropertyOptions{
			Class: class,
			Getter: func(val *TypeValue) (symbols.ValueObject, error) {
				return val.data[name], nil
			},
			Setter: func(val *TypeValue, newPropertyValue symbols.ValueObject) error {
				val.data[name] = newPropertyValue
				return nil
			},
		})
	}
	return &symbols.ClassDescriptors{
		Name: tc.Name,
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(symbols.Map, func(val *symbols.MapValue) (*TypeValue, error) {
				return &TypeValue{
					class: tc,
					data:  val.Map(),
				}, nil
			}),
		},
		Properties: propertyMap,
	}
}

type TypeValue struct {
	class TypeClass
	data  map[string]symbols.ValueObject
}

func (tv TypeValue) Class() symbols.Class {
	return tv.class
}
func (tv TypeValue) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range tv.data {
		out[k] = v.Value()
	}
	return out
}
