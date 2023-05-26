package stream

import (
	"fmt"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/symbols"
)

type Event struct {
	Topic      string
	ParentType symbols.Type
}

func (ev Event) ClassName() string {
	return ev.ParentType.ClassName()
}
func (ev Event) Fields() map[string]symbols.Class {
	return ev.ParentType.Fields()
}
func (ev Event) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddGenericConstructor(ev, func(fields map[string]symbols.ValueObject) (symbols.ValueObject, error) {
		return &EventInstance{ev, fields}, nil
	})
	return csMap
}
func (ev Event) Get(key string) (symbols.Object, error) {
	return nil, nil
}

func (ev Event) ObjectClassFromNode(ctx *context.Context, node ast.ContextObject) (symbols.Class, error) {
	assumedType, err := (context.TypeInterface{}).ObjectClassFromNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if typeClass, ok := assumedType.(symbols.Type); ok {
		return &Event{
			Topic:      fmt.Sprintf("%s.%s", ctx.Name, typeClass.Name),
			ParentType: typeClass,
		}, nil
	} else {
		panic("expected type class")
	}
}

func (ev *Event) Export() (symbols.Object, error) {
	return ev, nil
}

type EventInstance struct {
	ParentType Event
	Fields     map[string]symbols.ValueObject
}

func (ins EventInstance) Class() symbols.Class {
	return ins.ParentType
}
func (ins EventInstance) Value() interface{} {
	out := make(map[string]interface{})
	for key, obj := range ins.Fields {
		out[key] = obj.Value()
	}
	return out
}
func (ins *EventInstance) Set(key string, obj symbols.ValueObject) error {
	ins.Fields[key] = obj
	return nil
}
func (ins EventInstance) Get(key string) (symbols.Object, error) {
	return ins.Fields[key], nil
}
