package stream

import (
	"encoding/json"
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

type Query struct {
	Name    string
	Topic   string
	Private bool
	Comment string
	Handler *symbols.Function
	stream  *resource.NatsConnection `hash:"ignore"`
}

var QuerySignal = log.Signal("QUERY")
var QueryMessageSignal = log.Signal("QUERY_MESSAGE")

func (qr Query) ClassName() string {
	return "Query"
}
func (qr Query) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}
func (qr Query) Get(key string) (symbols.Object, error) {
	return nil, nil
}

func (qr *Query) Attach(process *runtime.Process) error {
	var conn resource.NatsConnection
	err := process.Resource("nats", &conn)
	if err != nil {
		return err
	}
	qr.stream = &conn
	return nil
}
func (qr *Query) AttachResource(process *runtime.Process) error {
	if qr.stream == nil {
		return fmt.Errorf("nats connection not initialized")
	}
	if qr.Handler.Handler != nil {
		qr.stream.Client.QueueSubscribe(qr.Topic, "handler_queue", func(m *nats.Msg) {
			start := time.Now()
			sendErrorResponse := func(streamError error) {
				bytes, err := json.Marshal(map[string]interface{}{
					"$error": streamError,
				})
				if err != nil {
					log.Output(log.LoggerMessage{
						LogLevel: log.LevelERROR,
						Signal:   QueryMessageSignal,
						Message:  "failed to marshal stream response when sending error",
						Data: map[string]string{
							"err":   err.Error(),
							"query": qr.Topic,
						},
					})
					return
				}
				err = m.Respond(bytes)
				if err != nil {
					log.Output(log.LoggerMessage{
						LogLevel: log.LevelERROR,
						Signal:   QueryMessageSignal,
						Message:  "failed to issue stream response when sending error",
						Data: map[string]string{
							"err":   err.Error(),
							"query": qr.Topic,
						},
					})
					return
				}
			}
			reportError := func(level log.LogLevel, logMessage string, data log.LogData, streamError error) {
				if data == nil {
					data = log.LogData{}
				}
				data["query"] = qr.Topic
				data["stream_response"] = streamError
				data["time"] = time.Since(start).String()
				log.Output(log.LoggerMessage{
					LogLevel: level,
					Signal:   QueryMessageSignal,
					Message:  logMessage,
					Data:     data,
				})
				sendErrorResponse(streamError)
			}

			var err error
			var generic symbols.ValueObject
			if len(qr.Arguments()) > 0 {
				generic, err = symbols.FromBytes(m.Data)
				if err != nil {
					reportError(log.LevelERROR, "failed to construct object from message", nil, symbols.Error{
						Name:    "MarshalError",
						Message: err.Error(),
					})
					return
				}
			}

			args := []symbols.ValueObject{}
			if generic != nil {
				args = append(args, generic)
			}

			result, err := qr.Handler.Call(args, Message{})
			if err != nil {
				if symErr, ok := err.(symbols.Error); ok {
					// not a runtime error -- an error returned from the invocation
					reportError(
						log.LevelINFO,
						fmt.Sprintf("\"%s\" failed", qr.Name),
						nil,
						symErr,
					)
				} else {
					reportError(
						log.LevelERROR,
						fmt.Sprintf("\"%s\" failed with unknown exception", qr.Name),
						log.LogData{"err": err.Error()},
						symbols.Error{
							Name:    "InternalError",
							Message: "unknown exception occured",
						},
					)
				}
				return
			}
			bytes := []byte("{}")
			if qr.Returns() != nil {
				bytes, err = json.Marshal(result.Value())
				if err != nil {
					reportError(
						log.LevelERROR,
						"failed to marshal response",
						log.LogData{"err": err.Error()},
						symbols.Error{
							Name:    "InternalError",
							Message: "failed to marshal response",
						},
					)
					return
				}
			}
			err = m.Respond(bytes)
			if err != nil {
				reportError(
					log.LevelERROR,
					"failed to send response",
					log.LogData{"err": err.Error()},
					symbols.Error{
						Name:    "InternalError",
						Message: "failed to send response",
					},
				)
				return
			}
			log.Printf(log.LevelINFO, QueryMessageSignal, "%s(): %s", qr.Name, time.Since(start))
		})
		log.Printf(log.LevelDEBUG, QuerySignal, "query \"%s\" listening for messages", qr.Name)
	}
	return nil
}
func (qr *Query) Detach() error {
	return nil
}

func (qr Query) Arguments() []symbols.Class {
	return qr.Handler.Arguments()
}
func (qr Query) Returns() symbols.Class {
	return qr.Handler.Returns()
}
func (qr Query) Call(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
	if qr.stream == nil {
		return nil, fmt.Errorf("nats connection not initialized")
	}
	var err error
	bytes := []byte{}
	if len(args) > 0 {
		bytes, err = json.Marshal(args[0].Value())
		if err != nil {
			return nil, err
		}
	}

	res, err := qr.stream.Client.Request(qr.Topic, bytes, time.Second*5)
	if err != nil {
		return nil, err
	}
	generic, err := symbols.FromBytes(res.Data)
	if err != nil {
		return nil, err
	}
	if obj, ok := generic.(*symbols.MapObject); ok {
		if err, ok := obj.Data["$error"]; ok {
			if errObject, ok := err.(*symbols.MapObject); ok {
				return nil, symbols.ErrorFromMapObject(errObject)
			} else {
				return nil, symbols.Error{
					Name:    "InternalError",
					Message: "unknown error was returned upstream",
				}
			}
		}
	}
	return symbols.Construct(qr.Returns(), generic)
}

func (Query) RemoteMethodClassFromNode(ctx *context.Context, node ast.RemoteContextMethod) (symbols.Class, error) {
	table := ctx.Symbols()
	args, returns, err := table.ResolveFunctionParameters(node.Parameters)
	if err != nil {
		return nil, err
	}
	if returns == nil {
		return nil, symbols.NodeError(node.Parameters, "query must have return type")
	}
	return &Query{
		Name:    node.Name,
		Topic:   fmt.Sprintf("%s.%s", ctx.Name, node.Name),
		Private: node.Private,
		Comment: node.Comment,
		Handler: &symbols.Function{
			ArgumentTypes: args,
			ReturnType:    returns,
			Handler:       nil,
		},
	}, nil
}
func (Query) MethodClassFromNode(ctx *context.Context, node ast.ContextMethod) (symbols.Class, error) {
	table := ctx.Symbols()
	fn, err := table.ResolveFunctionBlock(node.Block, Message{})
	if err != nil {
		return nil, err
	}
	if fn.ReturnType == nil {
		return nil, symbols.NodeError(node.Block, "query must have return type")
	}
	return &Query{
		Name:    node.Name,
		Topic:   fmt.Sprintf("%s.%s", ctx.Name, node.Name),
		Private: node.Private,
		Comment: node.Comment,
		Handler: fn,
	}, nil
}

func (qr Query) Export() (symbols.Object, error) {
	return qr, nil
}
