package interfaces

import (
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/interfaces/access"
	"github.com/hntrl/hyper/internal/interfaces/state"
	"github.com/hntrl/hyper/internal/interfaces/stream"
	"github.com/hntrl/hyper/internal/runtime"
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
