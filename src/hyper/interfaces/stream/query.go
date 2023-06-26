package stream

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hntrl/hyper/src/hyper/ast"
	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/hyper/symbols"
	"github.com/hntrl/hyper/src/hyper/symbols/errors"
	"github.com/hntrl/hyper/src/runtime/"
	"github.com/hntrl/hyper/src/runtime//log"
	"github.com/hntrl/hyper/src/runtime//resource"
	"github.com/nats-io/nats.go"
)

var QuerySignal = log.Signal("QUERY")
var QueryMessageSignal = log.Signal("QUERY_MESSAGE")

type QueryInterface struct{}

func (QueryInterface) FromNode(ctx *domain.Context, node ast.ContextMethod) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	if len(node.Block.Parameters.Arguments.Items) > 1 {
		return nil, errors.NodeError(node.Block.Parameters.Arguments, 0, "query must have only one argument")
	}
	if node.Block.Parameters.ReturnType == nil {
		return nil, errors.NodeError(node.Block.Parameters, 0, "query must have a return type")
	}
	table.Immutable["self"] = &symbols.ExpectedValueObject{Class: Message}
	fn, err := table.ResolveFunctionBlock(node.Block)
	if err != nil {
		return nil, err
	}
	query := Query{
		Name:        node.Name,
		Private:     node.Private,
		Comment:     node.Comment,
		Topic:       Topic(fmt.Sprintf("%s.%s", ctx.Identifier, node.Name)),
		PayloadType: nil,
		Returns:     fn.Returns(),
	}
	if len(fn.Arguments()) == 1 {
		query.PayloadType = fn.Arguments()[0]
	}
	consumer := QueryConsumer{
		query:   query,
		handler: fn,
	}
	if !node.Private {
		return &domain.ContextItem{
			HostItem:   consumer,
			RemoteItem: QueryEmitter{query: query},
		}, nil
	} else {
		return &domain.ContextItem{
			HostItem:   consumer,
			RemoteItem: nil,
		}, nil
	}
}

type Query struct {
	Name        string
	Private     bool
	Comment     string
	Topic       Topic
	PayloadType symbols.Class
	Returns     symbols.Class
}

type QueryConsumer struct {
	query   Query
	handler symbols.Callable
	stream  *resource.NatsConnection
}

func (consumer QueryConsumer) Arguments() []symbols.Class {
	if consumer.query.PayloadType == nil {
		return []symbols.Class{}
	}
	return []symbols.Class{consumer.query.PayloadType}
}
func (consumer QueryConsumer) Returns() symbols.Class {
	return consumer.query.Returns
}
func (consumer QueryConsumer) Call(args ...symbols.ValueObject) (symbols.ValueObject, error) {
	return consumer.handler.Call(args...)
}

func (consumer *QueryConsumer) Attach(process *runtime.Process) error {
	var conn resource.NatsConnection
	err := process.Resource("stream", &conn)
	if err != nil {
		return err
	}
	consumer.stream = &conn
	consumer.stream.Client.QueueSubscribe(string(consumer.query.Topic), "handler_queue", func(m *nats.Msg) {
		var payload symbols.ValueObject
		if consumer.query.PayloadType != nil {
			var err error
			payload, err = symbols.ValueFromBytes(m.Data)
			if err != nil {

			}
			payload, err = symbols.Construct(consumer.query.PayloadType, payload)
			if err != nil {

			}
		}
		result, err := consumer.handler.Call(payload)
		if err != nil {

		}
		bytes, err := json.Marshal(result.Value())
		if err != nil {

		}
		m.Respond(bytes)
	})
	return nil
}
func (consumer *QueryConsumer) Detach() error {
	consumer.stream = nil
	return nil
}

type QueryEmitter struct {
	query  Query
	stream *resource.NatsConnection
}

func (emitter QueryEmitter) Arguments() []symbols.Class {
	if emitter.query.PayloadType == nil {
		return []symbols.Class{}
	}
	return []symbols.Class{emitter.query.PayloadType}
}
func (emitter QueryEmitter) Returns() symbols.Class {
	return emitter.query.Returns
}
func (emitter QueryEmitter) Call(args ...symbols.ValueObject) (symbols.ValueObject, error) {
	if emitter.stream == nil {
		panic("stream connection not initialized")
	}
	bytes := []byte{}
	if emitter.query.PayloadType != nil {
		var err error
		bytes, err = json.Marshal(args[0].Value())
		if err != nil {
			return nil, err
		}
	}
	res, err := emitter.stream.Client.Request(string(emitter.query.Topic), bytes, time.Second*5)
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
				mappedError, err := symbols.Construct(symbols.Error, errMapValue)
				if err != nil {
					return nil, err
				}
				return nil, mappedError.(symbols.ErrorValue)
			}
			return nil, symbols.ErrorValue{
				Name:    "InternalError",
				Message: "unknown error was returned upstream",
			}
		}
	}
	return symbols.Construct(emitter.query.Returns, value)
}

func (emitter *QueryEmitter) Attach(process *runtime.Process) error {
	var conn resource.NatsConnection
	err := process.Resource("stream", &conn)
	if err != nil {
		return err
	}
	emitter.stream = &conn
	return nil
}
func (emitter *QueryEmitter) Detach() error {
	emitter.stream = nil
	return nil
}
