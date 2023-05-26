package state

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/symbols"
)

func RegisterDefaults(ctx *context.ContextBuilder, process *runtime.Process) {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	ctx.RegisterInterface("entity", &Entity{})
	ctx.RegisterInterface("projection", &Projection{})

	ctx.Selectors["deprecated_GenericID"] = symbols.NewFunction(symbols.FunctionOptions{
		Arguments: []symbols.Class{
			symbols.Integer{},
		},
		Returns: symbols.String{},
		Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
			return symbols.StringLiteral(strconv.Itoa(seededRand.Int())[0:args[0].(symbols.IntegerLiteral)]), nil
		},
	})
}
