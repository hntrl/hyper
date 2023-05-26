package interfaces

import (
	"fmt"
	"os"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/symbols"
)

type Parameter struct {
	ParentType   symbols.Class
	DefaultValue symbols.ValueObject
	Name         string
	DisplayName  string
	Description  string
}

func (pm Parameter) ClassName() string {
	return pm.ParentType.ClassName()
}
func (pm Parameter) Constructors() symbols.ConstructorMap {
	return pm.ParentType.Constructors()
}

func (pm Parameter) Class() symbols.Class {
	return pm.ParentType
}
func (pm Parameter) Value() interface{} {
	return pm.DefaultValue.Value()
}
func (pm Parameter) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, obj)
}
func (pm Parameter) Get(key string) (symbols.Object, error) {
	return pm.DefaultValue.Get(key)
}

// this is an extremely fucky way to do this -- but also it isn't?
func (pm Parameter) Attach(process *runtime.Process) error {
	return nil
}
func (pm Parameter) AttachResource(process *runtime.Process) error {
	varName := fmt.Sprintf("param_%s", pm.Name)
	val := os.Getenv(varName)
	if val == "" {
		return fmt.Errorf("param %s has no data: to fix, set env variable %s", pm.Name, varName)
	}
	generic, err := symbols.FromBytes([]byte(val))
	if err != nil {
		return fmt.Errorf("error initializing parameter %s: %s", pm.Name, err.Error())
	}
	obj, err := symbols.Construct(pm.ParentType, generic)
	if err != nil {
		return fmt.Errorf("error initializing parameter %s: %s", pm.Name, err.Error())
	}
	process.Context.Selectors[pm.Name] = obj
	return nil
}
func (pm Parameter) Detach() error {
	return nil
}

func (pm Parameter) ObjectClassFromNode(ctx *context.Context, node ast.ContextObject) (symbols.Class, error) {
	param := Parameter{Name: node.Name}
	table := ctx.Symbols()
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.AssignmentStatement:
			expr, err := table.ResolveValueObject(field.Init)
			if err != nil {
				return nil, err
			}
			switch field.Name {
			case "defaultValue":
				param.DefaultValue = expr
			case "name":
				if strLit, ok := expr.(symbols.StringLiteral); ok {
					param.DisplayName = string(strLit)
				} else {
					return nil, symbols.NodeError(field.Init, "%s not allowed in parameter", expr.Class().ClassName())
				}
			case "description":
				if strLit, ok := expr.(symbols.StringLiteral); ok {
					param.Description = string(strLit)
				} else {
					return nil, symbols.NodeError(field.Init, "%s not allowed in parameter", expr.Class().ClassName())
				}
			default:
				return nil, symbols.NodeError(field, "unknown assignment %s in parameter", field.Name)
			}
		case ast.TypeStatement:
			if field.Name == "type" {
				expr, err := table.ResolveTypeExpression(field.Init)
				if err != nil {
					return nil, err
				}
				param.ParentType = expr
			} else {
				return nil, symbols.NodeError(field, "unknown type %s in parameter", field.Name)
			}
		default:
			return nil, symbols.NodeError(field, "%T not allowed in parameter", field)
		}
	}
	// if param.Name == "" {
	// 	return nil, symbols.NodeError(node, "missing name in parameter")
	// }
	if param.ParentType == nil {
		return nil, symbols.NodeError(node, "missing type in parameter")
	}
	// if param.DefaultValue != nil {
	// 	err := symbols.CastObject(param.DefaultValue, &param.ParameterValue)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	ctx.Selectors[param.Name] = param.ParentType
	return param, nil
}
