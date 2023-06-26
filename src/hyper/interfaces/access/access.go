package access

import (
	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/runtime/"
)

func RegisterDefaults(builder *domain.ContextBuilder, process *runtime.Process) {
	builder.RegisterInterface("grant", GrantInterface{})
}
