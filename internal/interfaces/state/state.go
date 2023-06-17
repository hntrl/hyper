package state

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/symbols"
)

func RegisterDefaults(builder *domain.ContextBuilder, process *runtime.Process) {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	builder.RegisterInterface("entity", EntityInterface{})
	builder.RegisterInterface("projection", ProjectionInterface{})
	builder.RegisterSelector("unsafe_GenericID", symbols.NewFunction(symbols.FunctionOptions{
		Arguments: []symbols.Class{symbols.Integer},
		Returns:   symbols.String,
		Handler: func(len symbols.IntegerValue) (symbols.StringValue, error) {
			return symbols.StringValue(strconv.Itoa(seededRand.Int())[0:len]), nil
		},
	}))
}
