package stream

import (
	"github.com/hntrl/hyper/internal/interfaces/access"
	"github.com/hntrl/hyper/internal/symbols"
)

type Message struct {
	context MessageContext
	// data    symbols.ValueObject
}

func (msg Message) ClassName() string {
	return "Message"
}
func (msg Message) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}

func (msg Message) Class() symbols.Class {
	return msg
}
func (msg Message) Value() interface{} {
	return nil
}
func (msg Message) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, obj)
}
func (msg Message) Get(key string) (symbols.Object, error) {
	switch key {
	case "ctx":
		return msg.context, nil
	case "guard":
		return msg.context.user.Get("hasGrant")
	}
	return nil, nil
}

type MessageContext struct {
	user UserContext
}

func (ctx MessageContext) ClassName() string {
	return "MessageContext"
}
func (ctx MessageContext) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}

func (ctx MessageContext) Class() symbols.Class {
	return ctx
}
func (ctx MessageContext) Value() interface{} {
	return nil
}
func (ctx MessageContext) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, obj)
}
func (ctx MessageContext) Get(key string) (symbols.Object, error) {
	switch key {
	case "user":
		return ctx.user, nil
	}
	return nil, nil
}

type UserContext struct{}

func (uc UserContext) ClassName() string {
	return "UserContext"
}
func (uc UserContext) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}

func (uc UserContext) Class() symbols.Class {
	return uc
}
func (uc UserContext) Value() interface{} {
	return nil
}
func (uc UserContext) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, obj)
}
func (uc UserContext) Get(key string) (symbols.Object, error) {
	switch key {
	case "hasGrant":
		return symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				access.Grant{},
			},
			Returns: symbols.Boolean{},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				return symbols.BooleanLiteral(false), nil
			},
		}), nil
	}
	return nil, nil
}
