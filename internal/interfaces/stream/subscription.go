package stream

import (
	"fmt"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/symbols/errors"
	"github.com/nats-io/nats.go"
)

var SubscriptionSignal = log.Signal("SUBSCRIPTION")
var SubscriptionEventSignal = log.Signal("SUBSCRIPTION_EVENT")

type SubscriptionInterface struct{}

func (SubscriptionInterface) FromNode(ctx *domain.Context, node ast.ContextMethod) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	if node.Private {
		return nil, errors.NodeError(node, 0, "subscription cannot be private: subscriptions aren't exported")
	}
	if len(node.Block.Parameters.Arguments.Items) != 1 {
		return nil, errors.NodeError(node.Block.Parameters.Arguments, 0, "subscription must have exactly one argument")
	}
	if node.Block.Parameters.ReturnType != nil {
		return nil, errors.NodeError(node.Block.Parameters, 0, "subscription cannot have a return type")
	}
	fn, err := table.ResolveFunctionBlock(node.Block)
	if err != nil {
		return nil, err
	}
	event, ok := fn.Arguments()[0].(Event)
	if !ok {
		return nil, errors.NodeError(node.Block.Parameters.Arguments.Items[0], 0, "subscription argument must be an event")
	}
	sub := Subscription{
		Name:    node.Name,
		Private: node.Private,
		Comment: node.Comment,
		Topic:   Topic(fmt.Sprintf("%s.%s", ctx.Identifier, node.Name)),
		Event:   event,
	}
	return &domain.ContextItem{
		HostItem: SubscriptionConsumer{
			sub:     sub,
			handler: fn,
		},
		RemoteItem: nil,
	}, nil
}

type Subscription struct {
	Name    string
	Private bool
	Comment string
	Topic   Topic
	Event   Event
}

type SubscriptionConsumer struct {
	sub     Subscription
	handler symbols.Callable
	stream  *resource.NatsConnection
}

func (consumer *SubscriptionConsumer) Attach(process *runtime.Process) error {
	var conn resource.NatsConnection
	err := process.Resource("stream", &conn)
	if err != nil {
		return err
	}
	consumer.stream = &conn
	consumer.stream.Client.QueueSubscribe(string(consumer.sub.Topic), "subscription_queue", func(m *nats.Msg) {
		value, err := symbols.ValueFromBytes(m.Data)
		if err != nil {

		}
		payload, err := symbols.Construct(consumer.sub.Event, value)
		if err != nil {

		}
		_, err = consumer.handler.Call(payload)
		if err != nil {

		}
	})
	return nil
}
func (consumer *SubscriptionConsumer) Detach() error {
	consumer.stream = nil
	return nil
}
