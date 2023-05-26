package runtime

// RuntimeNode represents anything that requires shared resources to function
type RuntimeNode interface {
	Attach(*Process) error
	Detach() error
}

// RuntimeNode represents anything that should be attached to the runtime
// process (only executed for nodes in the root context)
type RuntimeResource interface {
	RuntimeNode
	AttachResource(*Process) error
}
