package stream

import (
	"github.com/hntrl/hyper/src/hyper/interfaces/access"
	"github.com/hntrl/hyper/src/hyper/symbols"
)

var (
	Message            = MessageClass{}
	MessageDescriptors = &symbols.ClassDescriptors{
		Name: "Message",
		Properties: symbols.ClassPropertyMap{
			"ctx": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: MessageContext,
				Getter: func(message MessageValue) (MessageContextValue, error) {
					return message.context, nil
				},
			}),
		},
	}
)

type MessageClass struct{}

func (MessageClass) Descriptors() *symbols.ClassDescriptors {
	return MessageDescriptors
}

type MessageValue struct {
	context MessageContextValue
}

func (msg MessageValue) Class() symbols.Class {
	return Message
}
func (msg MessageValue) Value() interface{} {
	return nil
}

var (
	MessageContext            = MessageContextClass{}
	MessageContextDescriptors = &symbols.ClassDescriptors{
		Name: "MessageContext",
		Properties: symbols.ClassPropertyMap{
			"user": symbols.PropertyAttributes(symbols.PropertyOptions{
				Class: UserContext,
				Getter: func(context MessageContextValue) (UserContextValue, error) {
					return context.user, nil
				},
			}),
		},
	}
)

type MessageContextClass struct{}

func (MessageContextClass) Descriptors() *symbols.ClassDescriptors {
	return MessageContextDescriptors
}

type MessageContextValue struct {
	user UserContextValue
}

func (ctx MessageContextValue) Class() symbols.Class {
	return MessageContext
}
func (ctx MessageContextValue) Value() interface{} {
	return nil
}

var (
	UserContext            = UserContextClass{}
	UserContextDescriptors = &symbols.ClassDescriptors{
		Name: "UserContext",
		Prototype: symbols.ClassPrototypeMap{
			"hasGrant": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class: UserContext,
				Arguments: []symbols.Class{
					access.Grant,
				},
				Returns: symbols.Boolean,
				Handler: func(user UserContextValue, grant access.GrantValue) (symbols.BooleanValue, error) {
					return symbols.BooleanValue(false), nil
				},
			}),
		},
	}
)

type UserContextClass struct{}

func (UserContextClass) Descriptors() *symbols.ClassDescriptors {
	return UserContextDescriptors
}

type UserContextValue struct {
}

func (ctx UserContextValue) Class() symbols.Class {
	return UserContext
}
func (ctx UserContextValue) Value() interface{} {
	return nil
}
