package interfaces

import (
	"fmt"
	"os"

	"github.com/hntrl/hyper/src/hyper/ast"
	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/hyper/symbols"
	"github.com/hntrl/hyper/src/hyper/symbols/errors"
	"github.com/hntrl/hyper/src/runtime/"
)

type ParameterInterface struct{}

func (ParameterInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	if node.Private {
		return nil, errors.NodeError(node, 0, "parameter cannot be private: parameters aren't exported")
	}
	param := Parameter{
		Name:    node.Name,
		Comment: node.Comment,
	}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.FieldAssignmentExpression:
			switch field.Name {
			case "defaultValue":
				defaultValue, err := table.ResolveExpression(field.Init)
				if err != nil {
					return nil, err
				}
				param.defaultValue = defaultValue
			default:
				return nil, errors.NodeError(field, 0, "unrecognized assignment %s in parameter", field.Name)
			}
		case ast.FieldExpression:
			switch field.Name {
			case "type":
				typeClass, err := table.EvaluateTypeExpression(field.Init)
				if err != nil {
					return nil, err
				}
				param.parentType = typeClass
			default:
				return nil, errors.NodeError(field, 0, "unrecognized field %s in parameter", field.Name)
			}
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in parameter", item)
		}
	}
	if param.parentType == nil {
		return nil, errors.NodeError(node, 0, "missing type in parameter")
	}
	if param.defaultValue != nil {
		obj, err := symbols.Construct(param.parentType, param.defaultValue)
		if err != nil {
			return nil, errors.NodeError(node, 0, "default value cannot be constructed: %s", err.Error())
		}
		param.defaultValue = obj
	}
	return &domain.ContextItem{
		HostItem:   param,
		RemoteItem: nil,
	}, nil
}

type Parameter struct {
	Name         string
	Comment      string
	parentType   symbols.Class
	defaultValue symbols.ValueObject
}

func (pm Parameter) Class() symbols.Class {
	return pm.parentType
}
func (pm Parameter) Value() interface{} {
	return nil
}

func (pm Parameter) Attach(process *runtime.Process) error {
	varName := fmt.Sprintf("param_%s", pm.Name)
	rawValue := os.Getenv(varName)
	if rawValue == "" {
		if pm.defaultValue != nil {
			process.Context.Selectors[pm.Name] = pm.defaultValue
			return nil
		}
		return fmt.Errorf("param %s has no data: to fix, set env variable %s", pm.Name, varName)
	}
	value, err := symbols.ValueFromBytes([]byte(rawValue))
	if err != nil {
		return fmt.Errorf("error initializing parameter %s: %s", pm.Name, err.Error())
	}
	constructedValue, err := symbols.Construct(pm.parentType, value)
	if err != nil {
		return fmt.Errorf("error initializing parameter %s: %s", pm.Name, err.Error())
	}
	process.Context.Selectors[pm.Name] = constructedValue
	return nil
}
func (pm Parameter) Detach() error {
	return nil
}
