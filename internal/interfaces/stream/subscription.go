package stream

import (
	"fmt"
	"time"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/nats-io/nats.go"
)

type Subscription struct {
	Name      string
	Event     Event
	Private   bool
	Comment   string
	Handler   *symbols.Function
	queueName string
}

var SubscriptionSignal = log.Signal("SUBSCRIPTION")
var SubscriptionEventSignal = log.Signal("SUBSCRIPTION_EVENT")

func (sb Subscription) ClassName() string {
	return "Subscription"
}
func (sb Subscription) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}
func (sb Subscription) Get(key string) (symbols.Object, error) {
	return nil, nil
}

func (sb *Subscription) Attach(process *runtime.Process) error {
	return nil
}
func (sb *Subscription) AttachResource(process *runtime.Process) error {
	var conn resource.NatsConnection
	err := process.Resource("nats", &conn)
	if err != nil {
		return err
	}
	if sb.Handler != nil {
		if _, err := conn.Client.QueueSubscribe(sb.Event.Topic, sb.queueName, func(m *nats.Msg) {
			start := time.Now()
			reportError := func(level log.LogLevel, err error) {
				log.Output(log.LoggerMessage{
					LogLevel: level,
					Signal:   SubscriptionEventSignal,
					Message:  fmt.Sprintf("%s() failed in %s", sb.Name, time.Since(start)),
					Data: log.LogData{
						"err":          err.Error(),
						"subscription": sb.Name,
					},
				})
			}

			obj, err := symbols.FromBytes(m.Data)
			if err != nil {
				reportError(log.LevelERROR, err)
				return
			}
			_, err = sb.Handler.Call([]symbols.ValueObject{obj}, Message{})
			if err != nil {
				reportError(log.LevelERROR, err)
				return
			}
			log.Printf(log.LevelINFO, SubscriptionEventSignal, "%s(%s): %s", sb.Name, sb.Event.Topic, time.Since(start))
		}); err != nil {
			return err
		}
		log.Printf(log.LevelDEBUG, SubscriptionSignal, "subscription \"%s\" listening for \"%s\" events", sb.Name, sb.Event.Topic)
	}
	return nil
}
func (sb *Subscription) Detach() error {
	return nil
}

func validateSubscriptionParameters(node ast.FunctionParameters, args []symbols.Class, returns symbols.Class) error {
	if len(args) != 1 {
		return symbols.NodeError(node.Arguments.Items[0], "subscription Handler must have exactly one argument")
	}
	if returns != nil {
		return symbols.NodeError(node.ReturnType, "subscription Handler return type is ambiguous")
	}
	if _, ok := args[0].(*Event); !ok {
		return symbols.NodeError(node.Arguments.Items[0], "subscription Handler argument must be an event")
	}
	return nil
}

func (Subscription) RemoteMethodClassFromNode(ctx *context.Context, node ast.RemoteContextMethod) (symbols.Class, error) {
	symbols := ctx.Symbols()
	args, returns, err := symbols.ResolveFunctionParameters(node.Parameters)
	if err != nil {
		return nil, err
	}
	err = validateSubscriptionParameters(node.Parameters, args, returns)
	if err != nil {
		return nil, err
	}
	event := args[0].(*Event)
	return &Subscription{
		Name:    node.Name,
		Event:   *event,
		Private: node.Private,
		Comment: node.Comment,
	}, nil
}
func (Subscription) MethodClassFromNode(ctx *context.Context, node ast.ContextMethod) (symbols.Class, error) {
	table := ctx.Symbols()
	fn, err := table.ResolveFunctionBlock(node.Block, Message{})
	if err != nil {
		return nil, err
	}

	err = validateSubscriptionParameters(node.Block.Parameters, fn.Arguments(), fn.Returns())
	if err != nil {
		return nil, err
	}
	event := fn.ArgumentTypes[0].(*Event)
	return &Subscription{
		Name:      node.Name,
		Event:     *event,
		Private:   node.Private,
		Comment:   node.Comment,
		Handler:   fn,
		queueName: fmt.Sprintf("%s.%s:handler_queue", ctx.Name, node.Name),
	}, nil
}
