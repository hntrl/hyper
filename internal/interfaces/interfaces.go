package interfaces

import (
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/interfaces/access"
	"github.com/hntrl/hyper/internal/interfaces/state"
	"github.com/hntrl/hyper/internal/interfaces/stream"
	"github.com/hntrl/hyper/internal/runtime"
)

func RegisterDefaults(ctx *context.ContextBuilder, process *runtime.Process) {
	stream.RegisterDefaults(ctx, process)
	state.RegisterDefaults(ctx, process)
	ctx.RegisterInterface("grant", access.Grant{})
	ctx.RegisterInterface("template", &Template{})
	ctx.RegisterInterface("enum", Enum{})
	ctx.RegisterInterface("file", FileType{})
	ctx.RegisterInterface("param", Parameter{})
}
