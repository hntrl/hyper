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

type Command struct {
	Name    string
	Topic   string
	Private bool
	Comment string
	Handler *symbols.Function
	stream  *resource.NatsConnection `hash:"ignore"`
}

var CommandSignal = log.Signal("COMMAND")
var CommandMessageSignal = log.Signal("COMMAND_MESSAGE")

func (cmd Command) ClassName() string {
	return "Command"
}
func (cmd Command) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}
func (cmd Command) Get(key string) (symbols.Object, error) {
	return nil, nil
}

func (cmd *Command) Attach(process *runtime.Process) error {
	var conn resource.NatsConnection
	err := process.Resource("nats", &conn)
	if err != nil {
		return err
	}
	cmd.stream = &conn
	return nil
}
func (cmd *Command) AttachResource(process *runtime.Process) error {
	if cmd.stream == nil {
		return fmt.Errorf("nats connection not initialized")
	}
	if cmd.Handler.Handler != nil {
		cmd.stream.Client.QueueSubscribe(cmd.Topic, "handler_queue", func(m *nats.Msg) {
			start := time.Now()
			sendErrorResponse := func(streamError error) {
				bytes, err := json.Marshal(map[string]interface{}{
					"$error": streamError,
				})
				if err != nil {
					log.Output(log.LoggerMessage{
						LogLevel: log.LevelERROR,
						Signal:   CommandMessageSignal,
						Message:  "failed to marshal stream response when sending error",
						Data: map[string]string{
							"err":     err.Error(),
							"command": cmd.Topic,
						},
					})
					return
				}
				err = m.Respond(bytes)
				if err != nil {
					log.Output(log.LoggerMessage{
						LogLevel: log.LevelERROR,
						Signal:   CommandMessageSignal,
						Message:  "failed to issue stream response when sending error",
						Data: map[string]string{
							"err":     err.Error(),
							"command": cmd.Topic,
						},
					})
					return
				}
			}
			reportError := func(level log.LogLevel, logMessage string, data log.LogData, streamError error) {
				if data == nil {
					data = log.LogData{}
				}
				data["command"] = cmd.Topic
				data["stream_response"] = streamError
				data["time"] = time.Since(start).String()
				log.Output(log.LoggerMessage{
					LogLevel: level,
					Signal:   CommandMessageSignal,
					Message:  logMessage,
					Data:     data,
				})
				sendErrorResponse(streamError)
			}

			var err error
			var generic symbols.ValueObject
			if len(cmd.Arguments()) > 0 {
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

			result, err := cmd.Handler.Call(args, Message{})
			if err != nil {
				if symErr, ok := err.(symbols.Error); ok {
					// not a runtime error -- an error returned from the invocation
					reportError(
						log.LevelINFO,
						fmt.Sprintf("\"%s\" failed", cmd.Name),
						nil,
						symErr,
					)
				} else {
					reportError(
						log.LevelERROR,
						fmt.Sprintf("\"%s\" failed with unknown exception", cmd.Name),
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
			if cmd.Returns() != nil {
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
			log.Printf(log.LevelINFO, CommandMessageSignal, "%s(): %s", cmd.Name, time.Since(start))
		})
		log.Printf(log.LevelDEBUG, CommandSignal, "command \"%s\" listening for messages", cmd.Name)
	}
	return nil
}
func (cmd *Command) Detach() error {
	return nil
}

func (cmd Command) Arguments() []symbols.Class {
	return cmd.Handler.Arguments()
}
func (cmd Command) Returns() symbols.Class {
	return cmd.Handler.Returns()
}
func (cmd Command) Call(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
	if cmd.stream == nil {
		return nil, fmt.Errorf("nats connection not initialized")
	}

	bytes, err := json.Marshal(args[0].Value())
	if err != nil {
		return nil, err
	}
	res, err := cmd.stream.Client.Request(cmd.Topic, bytes, time.Second*5)
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
	if cmd.Returns() == nil {
		return symbols.NilLiteral{}, nil
	}
	return symbols.Construct(cmd.Returns(), generic)
}

func (Command) RemoteMethodClassFromNode(ctx *context.Context, node ast.RemoteContextMethod) (symbols.Class, error) {
	table := ctx.Symbols()
	args, returns, err := table.ResolveFunctionParameters(node.Parameters)
	if err != nil {
		return nil, err
	}
	return &Command{
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
func (Command) MethodClassFromNode(ctx *context.Context, node ast.ContextMethod) (symbols.Class, error) {
	symbols := ctx.Symbols()
	fn, err := symbols.ResolveFunctionBlock(node.Block, Message{})
	if err != nil {
		return nil, err
	}
	return &Command{
		Name:    node.Name,
		Topic:   fmt.Sprintf("%s.%s", ctx.Name, node.Name),
		Private: node.Private,
		Comment: node.Comment,
		Handler: fn,
	}, nil
}

func (cmd Command) Export() (symbols.Object, error) {
	return cmd, nil
}
