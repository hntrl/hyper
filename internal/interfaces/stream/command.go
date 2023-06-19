package stream

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/symbols/errors"
	"github.com/nats-io/nats.go"
)

var CommandSignal = log.Signal("COMMAND")
var CommandMessageSignal = log.Signal("COMMAND_MESSAGE")

type CommandInterface struct{}

func (CommandInterface) FromNode(ctx *domain.Context, node ast.ContextMethod) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	if len(node.Block.Parameters.Arguments.Items) > 1 {
		return nil, errors.NodeError(node.Block.Parameters.Arguments, 0, "command must have only one argument")
	}
	table.Immutable["self"] = &symbols.ExpectedValueObject{Class: Message}
	fn, err := table.ResolveFunctionBlock(node.Block)
	if err != nil {
		return nil, err
	}
	cmd := Command{
		Name:        node.Name,
		Private:     node.Private,
		Comment:     node.Comment,
		Topic:       Topic(fmt.Sprintf("%s.%s", ctx.Identifier, node.Name)),
		PayloadType: nil,
		Returns:     fn.Returns(),
	}
	if len(fn.Arguments()) == 1 {
		cmd.PayloadType = fn.Arguments()[0]
	}
	consumer := CommandConsumer{
		cmd:     cmd,
		handler: fn,
	}
	if !node.Private {
		return &domain.ContextItem{
			HostItem:   consumer,
			RemoteItem: CommandEmitter{cmd: cmd},
		}, nil
	} else {
		return &domain.ContextItem{
			HostItem:   consumer,
			RemoteItem: nil,
		}, nil
	}
}

type Command struct {
	Name        string
	Private     bool
	Comment     string
	Topic       Topic
	PayloadType symbols.Class
	Returns     symbols.Class
}

// CommandConsumer represents the abstraction used by the runtime to attach to a stream and process incoming messages on behalf of a Command.
type CommandConsumer struct {
	cmd     Command
	handler symbols.Callable
	stream  *resource.NatsConnection
}

func (consumer CommandConsumer) Arguments() []symbols.Class {
	if consumer.cmd.PayloadType == nil {
		return []symbols.Class{}
	}
	return []symbols.Class{consumer.cmd.PayloadType}
}
func (consumer CommandConsumer) Returns() symbols.Class {
	return consumer.cmd.Returns
}
func (consumer CommandConsumer) Call(args ...symbols.ValueObject) (symbols.ValueObject, error) {
	return consumer.handler.Call(args...)
}

func (consumer *CommandConsumer) Attach(process runtime.Process) error {
	consumer.stream.Client.QueueSubscribe(string(consumer.cmd.Topic), "handler_queue", func(m *nats.Msg) {
		var payload symbols.ValueObject
		if consumer.cmd.PayloadType != nil {
			value, err := symbols.ValueFromBytes(m.Data)
			if err != nil {

			}
			payload, err = symbols.Construct(consumer.cmd.PayloadType, value)
			if err != nil {

			}
		}
		result, err := consumer.handler.Call(payload)
		if err != nil {

		}
		if consumer.cmd.Returns != nil {
			bytes, err := json.Marshal(result.Value())
			if err != nil {

			}
			m.Respond(bytes)
		}
	})
	return nil
}
func (consumer *CommandConsumer) Detach(process runtime.Process) error {
	consumer.stream = nil
	return nil
}

type CommandEmitter struct {
	cmd    Command
	stream *resource.NatsConnection
}

func (emitter CommandEmitter) Arguments() []symbols.Class {
	if emitter.cmd.PayloadType == nil {
		return []symbols.Class{}
	}
	return []symbols.Class{emitter.cmd.PayloadType}
}
func (emitter CommandEmitter) Returns() symbols.Class {
	return emitter.cmd.Returns
}
func (emitter CommandEmitter) Call(args ...symbols.ValueObject) (symbols.ValueObject, error) {
	if emitter.stream == nil {
		panic("stream connection not initialized")
	}
	bytes := []byte{}
	if emitter.cmd.PayloadType != nil {
		var err error
		bytes, err = json.Marshal(args[0].Value())
		if err != nil {
			return nil, err
		}
	}
	if emitter.cmd.Returns != nil {
		res, err := emitter.stream.Client.Request(string(emitter.cmd.Topic), bytes, time.Second*5)
		if err != nil {
			return nil, err
		}
		value, err := symbols.ValueFromBytes(res.Data)
		if err != nil {
			return nil, err
		}
		if mapValue, ok := value.(*symbols.MapValue); ok {
			if errValue := mapValue.Get("$error"); errValue != nil {
				if errMapValue, ok := errValue.(*symbols.MapValue); ok {
					constructedErr, err := symbols.Construct(symbols.Error, errMapValue)
					if err != nil {
						return nil, err
					}
					return nil, constructedErr.(symbols.ErrorValue)
				}
				return nil, symbols.ErrorValue{
					Name:    "InternalError",
					Message: "unknown error was returned upstream",
				}
			}
		}
		return symbols.Construct(emitter.cmd.Returns, value)
	} else {
		err := emitter.stream.Client.Publish(string(emitter.cmd.Topic), bytes)
		return nil, err
	}
}

func (emitter *CommandEmitter) Attach(process runtime.Process) error {
	var conn resource.NatsConnection
	err := process.Resource("stream", &conn)
	if err != nil {
		return err
	}
	emitter.stream = &conn
	return nil
}
func (emitter *CommandEmitter) Detach(process runtime.Process) error {
	emitter.stream = nil
	return nil
}
