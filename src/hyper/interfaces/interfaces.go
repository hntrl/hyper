package interfaces

import (
	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/hyper/interfaces/access"
	"github.com/hntrl/hyper/src/hyper/interfaces/state"
	"github.com/hntrl/hyper/src/hyper/interfaces/stream"
	"github.com/hntrl/hyper/src/runtime/"
)

func RegisterDefaults(builder *domain.ContextBuilder, process *runtime.Process) {
	access.RegisterDefaults(builder, process)
	state.RegisterDefaults(builder, process)
	stream.RegisterDefaults(builder, process)
	builder.RegisterInterface("enum", EnumInterface{})
	builder.RegisterInterface("file", FileInterface{})
	builder.RegisterInterface("param", ParameterInterface{})
	builder.RegisterInterface("template", TemplateInterface{})
	builder.RegisterInterface("type", TypeInterface{})
}
