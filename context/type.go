package context

import (
	"fmt"

	"github.com/hntrl/lang/language/nodes"
	"github.com/hntrl/lang/symbols"
)

type TypeInterface struct{}

func (ti TypeInterface) ObjectClassFromNode(ctx *Context, node nodes.ContextObject) (symbols.Class, error) {
	table := ctx.Symbols()

	t := symbols.Type{
		Name:       node.Name,
		Private:    node.Private,
		Comment:    node.Comment,
		Properties: make(map[string]symbols.Class),
	}

	if node.Extends != nil {
		extendsType := nodes.TypeExpression{IsArray: false, IsPartial: false, IsOptional: false, Selector: *node.Extends}
		class, err := table.ResolveTypeExpression(extendsType)
		if err != nil {
			return nil, err
		}
		objectClass, ok := class.(symbols.ObjectClass)
		if !ok {
			return nil, fmt.Errorf("cannot extend %s", class.ClassName())
		}
		if fields := objectClass.Fields(); fields != nil {
			for k, v := range fields {
				t.Properties[k] = v
			}
		}
	}
	for _, item := range node.Fields {
		typeExpr, ok := item.Init.(nodes.TypeStatement)
		if !ok {
			return nil, fmt.Errorf("expected type statement")
		}
		obj, err := table.ResolveTypeExpression(typeExpr.Init)
		if err != nil {
			return nil, err
		}
		t.Properties[typeExpr.Name] = obj
	}

	return t, nil
}
