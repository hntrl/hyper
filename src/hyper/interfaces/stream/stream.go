package stream

import (
	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/runtime/"
)

type Topic string

func RegisterDefaults(builder *domain.ContextBuilder, process *runtime.Process) {
	builder.RegisterInterface("command", CommandInterface{})
	builder.RegisterInterface("event", EventInterface{})
	builder.RegisterInterface("query", QueryInterface{})
	builder.RegisterInterface("sub", SubscriptionInterface{})
	builder.RegisterSelector("emit", makeEventEmitterFunction(process))
}
