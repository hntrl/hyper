package stream

import (
	"encoding/json"
	"fmt"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/symbols/errors"
)

type EventInterface struct{}

func (EventInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	ev := Event{
		Name:       node.Name,
		Private:    node.Private,
		Comment:    node.Comment,
		Topic:      Topic(fmt.Sprintf("%s.%s", ctx.Identifier, node.Name)),
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
			ev.Properties[k] = v.PropertyClass
		}
	}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.FieldExpression:
			class, err := table.EvaluateTypeExpression(field.Init)
			if err != nil {
				return nil, err
			}
			ev.Properties[field.Name] = class
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in type", item)
		}
	}
	if !node.Private {
		return &domain.ContextItem{
			HostItem:   ev,
			RemoteItem: ev,
		}, nil
	} else {
		return &domain.ContextItem{
			HostItem:   ev,
			RemoteItem: nil,
		}, nil
	}
}

type Event struct {
	Name       string
	Private    bool
	Comment    string
	Topic      Topic
	Properties map[string]symbols.Class
}

func (ev Event) Descriptors() *symbols.ClassDescriptors {
	propertyDescriptors := make(symbols.ClassPropertyMap)
	for name, class := range ev.Properties {
		propertyDescriptors[name] = symbols.PropertyAttributes(symbols.PropertyOptions{
			Class: class,
			Getter: func(val *EventObject) (symbols.ValueObject, error) {
				return val.data[name], nil
			},
		})
	}
	return &symbols.ClassDescriptors{
		Name: ev.Name,
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(symbols.Map, func(val *symbols.MapValue) (EventObject, error) {
				return EventObject{
					parentType: ev,
					data:       val.Map(),
				}, nil
			}),
		},
		Properties: propertyDescriptors,
	}
}

type EventObject struct {
	parentType Event
	data       map[string]symbols.ValueObject
}

func (ev EventObject) Class() symbols.Class {
	return ev.parentType
}
func (ev EventObject) Value() interface{} {
	return ev.data
}

func makeEventEmitterFunction(process *runtime.Process) symbols.Callable {
	return symbols.NewFunction(symbols.FunctionOptions{
		Arguments: []symbols.Class{
			// FIXME: same problem as len(), but instead of indexable it should be
			// classes.EventInstance
			// symbols.Any{},
		},
		Returns: nil,
		Handler: func(ev symbols.ValueObject) (symbols.ValueObject, error) {
			eventObject, ok := ev.(*EventObject)
			if !ok {
				return nil, fmt.Errorf("cannot emit non-event")
			}
			var conn resource.NatsConnection
			err := process.Resource("nats", &conn)
			if err != nil {
				return nil, err
			}
			bytes, err := json.Marshal(eventObject.Value())
			if err != nil {
				return nil, err
			}
			err = conn.Client.Publish(string(eventObject.parentType.Topic), bytes)
			if err != nil {
				return nil, err
			}
			log.Printf(log.LevelINFO, log.Signal("EVENT"), "\"%s\" emitted", eventObject.parentType.Topic)

			return nil, nil
		},
	})
}
