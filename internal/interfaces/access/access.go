package access

import (
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/runtime"
)

func RegisterDefaults(builder *domain.ContextBuilder, process *runtime.Process) {
	builder.RegisterInterface("grant", GrantInterface{})
}
