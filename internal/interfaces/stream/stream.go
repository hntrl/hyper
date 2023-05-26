package stream

import (
	"encoding/json"
	"fmt"

	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"
)

func RegisterDefaults(ctx *context.ContextBuilder, process *runtime.Process) {
	ctx.RegisterInterface("command", &Command{})
	ctx.RegisterInterface("event", &Event{})
	ctx.RegisterInterface("query", &Query{})
	ctx.RegisterInterface("sub", &Subscription{})

	ctx.Selectors["emit"] = symbols.NewFunction(symbols.FunctionOptions{
		Arguments: []symbols.Class{
			// FIXME: same problem as len(), but instead of indexable it should be
			// classes.EventInstance
			symbols.AnyClass{},
		},
		Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
			eventInstance, ok := args[0].(*EventInstance)
			if !ok {
				return nil, fmt.Errorf("cannot emit non-event")
			}

			var conn resource.NatsConnection
			err := process.Resource("nats", &conn)
			if err != nil {
				return nil, err
			}

			bytes, err := json.Marshal(args[0].Value())
			if err != nil {
				return nil, err
			}

			err = conn.Client.Publish(eventInstance.ParentType.Topic, bytes)
			if err != nil {
				return nil, err
			}
			log.Printf(log.LevelINFO, log.Signal("EVENT"), "\"%s\" emitted", eventInstance.ParentType.Topic)
			return nil, nil
		},
	})
}
