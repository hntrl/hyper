package interfaces

import (
	"fmt"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/tokens"
)

type Enum struct {
	Name    string
	Private bool
	Comment string
	Items   map[string]EnumItem `hash:"ignore"`
}

func (en Enum) ClassName() string {
	return en.Name
}
func (en Enum) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddConstructor(en, func(obj symbols.ValueObject) (symbols.ValueObject, error) {
		return obj, nil
	})
	csMap.AddConstructor(symbols.String{}, func(obj symbols.ValueObject) (symbols.ValueObject, error) {
		for _, item := range en.Items {
			if item.StringValue == symbols.StringLiteral(obj.Value().(string)) {
				return item, nil
			}
		}
		if item, ok := en.Items[obj.Value().(string)]; ok {
			return item, nil
		}
		return nil, fmt.Errorf("%s not valid for %s", obj, en.ClassName())
	})
	return csMap
}
func (en Enum) Get(key string) (symbols.Object, error) {
	return en.Items[key], nil
}

func (en Enum) ComparableRules() symbols.ComparatorRules {
	rules := symbols.NewComparatorRules()
	rules.AddComparator(en, tokens.EQUALS, func(a, b symbols.ValueObject) (symbols.ValueObject, error) {
		return symbols.BooleanLiteral(a.Value() == b.Value()), nil
	})
	rules.AddComparator(en, tokens.NOT_EQUALS, func(a, b symbols.ValueObject) (symbols.ValueObject, error) {
		return symbols.BooleanLiteral(a.Value() != b.Value()), nil
	})
	return rules
}

func (en Enum) Export() (symbols.Object, error) {
	return en, nil
}

type EnumItem struct {
	ParentType  Enum
	StringValue symbols.StringLiteral `hash:"ignore"`
}

func (en EnumItem) Class() symbols.Class {
	return en.ParentType
}
func (en EnumItem) Value() interface{} {
	return en.StringValue.Value()
}
func (en EnumItem) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, en)
}
func (en EnumItem) Get(key string) (symbols.Object, error) {
	return nil, nil
}

func (en Enum) ObjectClassFromNode(ctx *context.Context, node ast.ContextObject) (symbols.Class, error) {
	obj := Enum{
		Name:    node.Name,
		Private: node.Private,
		Comment: node.Comment,
	}
	items := make(map[string]EnumItem)
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.EnumStatement:
			items[field.Name] = EnumItem{
				obj,
				symbols.StringLiteral(field.Init),
			}
		default:
			return nil, fmt.Errorf("parsing: %T not allowed in enum", item)
		}
	}
	obj.Items = items
	return obj, nil
}

// func (en Enum) IsExported() bool {
// 	return !en.Private
// }
// func (en Enum) ExportedObject() symbols.Object {
// 	return en
// }
